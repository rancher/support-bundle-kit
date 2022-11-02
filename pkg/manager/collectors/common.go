package collectors

import (
	"io"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/sirupsen/logrus"

	"github.com/rancher/support-bundle-kit/pkg/manager/client"
)

type encodeToYAMLFile func(obj interface{}, path string, errLog io.Writer)

type moduleCollector interface {
	generateYAMLs()
	toObj(b []byte, groupVersion, kind string, resources ...string) (interface{}, error)
}

func InitModuleCollector(moduleName string, yamlDir string, nameSpaces []string, discovery *client.DiscoveryClient, exclude client.ExcludeFilter, encodeFunc encodeToYAMLFile, errLog io.Writer) interface{} {
	common := NewCommonModule(discovery, encodeFunc, exclude, yamlDir, errLog)
	switch strings.ToLower(moduleName) {
	case "cluster":
		return NewClusterModule(common, "Cluster")
	case "default":
		return NewDefaultModule(common, "Default", nameSpaces)
	case "harvester":
		return NewHarvesterModule(common, "Harvester")
	default:
		return nil
	}
}

func GetAllSupportBundleYAMLs(modules []interface{}) {
	logrus.Infof("Prepare to get all support bundle yamls!")
	for _, raw_module := range modules {
		module := raw_module.(moduleCollector)
		module.generateYAMLs()
	}
}

type common struct {
	discovery  *client.DiscoveryClient
	encodeFunc encodeToYAMLFile
	exclude    client.ExcludeFilter
	yamlsDir   string
	errorLog   io.Writer
}

func NewCommonModule(discovery *client.DiscoveryClient, encodeFunc encodeToYAMLFile, exclude client.ExcludeFilter, YamlsDir string, ErrorLog io.Writer) *common {
	return &common{
		discovery:  discovery,
		encodeFunc: encodeFunc,
		exclude:    exclude,
		yamlsDir:   YamlsDir,
		errorLog:   ErrorLog,
	}
}

func (c common) toObjCommon(b []byte, groupVersion, kind string) (*gabs.Container, error) {
	re := regexp.MustCompile(`("[a-zA-Z]+":)(null,)`)
	replaceString := re.ReplaceAllString(string(b), "$1\"null\",")

	re = regexp.MustCompile(`(\\"[a-zA-Z]+\\":)(null,)`)
	replaceString = re.ReplaceAllString(replaceString, "$1\\\"null\\\",")

	finalString := strings.ReplaceAll(replaceString, `""`, `"null"`)
	jsonParsed, err := gabs.ParseJSON([]byte(finalString))
	if err != nil {
		logrus.Errorf("Unable to parse json: %s, %s", groupVersion, kind)
		return nil, err
	}
	// the yaml contains a list of resources
	if _, err = jsonParsed.SetP("List", "kind"); err != nil {
		logrus.Error("Unable to set kind for list.")
		return nil, err
	}

	if _, err = jsonParsed.SetP("v1", "apiVersion"); err != nil {
		logrus.Error("Unable to set apiVersion for list.")
		return nil, err
	}

	for _, child := range jsonParsed.S("items").Children() {
		if _, err = child.SetP(groupVersion, "apiVersion"); err != nil {
			logrus.Error("Unable to set apiVersion field.")
			return nil, err
		}

		if _, err = child.SetP(kind, "kind"); err != nil {
			logrus.Error("Unable to set kind field.")
			return nil, err
		}
	}

	if kind == "Secret" {
		secretsTargetData := getSecretsTargetData()
		for _, child := range jsonParsed.S("items").Children() {
			if exists := child.Exists("data"); exists {
				currentDataItems := child.S("data").Data().(map[string]interface{})
				newItems := make(map[string]interface{})
				for key, item := range currentDataItems {
					if _, exists := secretsTargetData[key]; !exists {
						continue
					}
					newItems[key] = item
				}
				_, err := child.SetP(newItems, "data")
				if err != nil {
					logrus.Error("Unable to clear data section")
				}
			}
		}
	}
	return jsonParsed, nil
}

func getSecretsTargetData() map[string]bool {
	dataKeys := map[string]bool{
		"applied-checksum":        true,
		"applied-output":          true,
		"applied-periodic-output": true,
		"failed-checksum":         true,
		"failed-output":           true,
		"failure-count":           true,
		"failure-threshold":       true,
		"last-apply-time":         true,
		"max-failures":            true,
		"probe-statuses":          true,
		"success-count":           true,
	}
	return dataKeys
}

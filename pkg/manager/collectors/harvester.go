package collectors

import (
	"fmt"
	"path/filepath"

	"github.com/Jeffail/gabs/v2"
	"github.com/rancher/wrangler/pkg/slice"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ignoreHarvesterSettingsList = []string{"cluster-registration-url", "containerd-registry", "additional-ca", "ssl-certificates"}

type harvesterModule struct {
	c    *common
	name string
}

func NewHarvesterModule(common *common, name string) *harvesterModule {
	return &harvesterModule{
		c:    common,
		name: name,
	}
}

func (module harvesterModule) generateYAMLs() {
	logrus.Infof("[%s] generate YAMLs, yamlsDir: %s", module.name, module.c.yamlsDir)

	extraResources := getHarvesterExtraResource()
	for namespace, resourceLists := range extraResources {
		dir := filepath.Join(module.c.yamlsDir, "namespaced", namespace)
		objs, err := module.c.discovery.SpecificResourcesForNamespace(module.toObj, module.name, namespace, resourceLists, module.c.errorLog)

		if err != nil {
			logrus.WithError(err).Error("Unable to fetch namespaced resources")
			_, _ = fmt.Fprintf(module.c.errorLog, "Unable to fetch namespaced resources: %v\n", err)
			return
		}

		for name, obj := range objs {
			file := filepath.Join(dir, name+".yaml")
			logrus.Debugf("Prepare to encode to yaml file path: %s", file)
			module.c.encodeFunc(obj, file, module.c.errorLog)
		}
	}

	dir := filepath.Join(module.c.yamlsDir, "cluster")
	objs, err := module.c.discovery.ResourcesForCluster(module.toClusterObj, module.skipClusterObjects, module.c.errorLog)

	if err != nil {
		logrus.WithError(err).Error("Unable to fetch cluster resources")
		_, _ = fmt.Fprintf(module.c.errorLog, "Unable to fetch cluster resources: %v\n", err)
		return
	}

	for name, obj := range objs {
		file := filepath.Join(dir, name+".yaml")
		logrus.Debugf("Prepare to encode to yaml file path: %s", file)
		module.c.encodeFunc(obj, file, module.c.errorLog)
	}

}

func (module harvesterModule) toObj(b []byte, groupVersion, kind string, resources ...string) (interface{}, error) {
	jsonParsed, err := module.c.toObjCommon(b, groupVersion, kind)

	if err != nil {
		return nil, err
	}

	/* Checking all resource and log the unknown one */
	for _, resource := range resources {
		switch resource {
		case "secrets":
			currentItems, _ := jsonParsed.S("items").Data().([]interface{})
			logrus.Debugf("Whole items: %v", currentItems)
			var newItems []interface{}
			for _, item := range currentItems {
				gItem := gabs.Wrap(item)
				if find := gItem.S("type").Data().(string) == "rke.cattle.io/machine-plan"; find {
					logrus.Debugf("Prepare to append item: %v", gItem.Data().(map[string]interface{}))
					newItems = append(newItems, item)
				}
			}
			if _, err := jsonParsed.Set(newItems, "items"); err != nil {
				return nil, err
			}
		default:
			// undefined resource, just logged it.
			logrus.Warnf("Could not handle unknown resource %s", resource)
		}
	}
	return jsonParsed.Data(), nil
}

/* return map{<namespace>: [resource]} */
func getHarvesterExtraResource() map[string][]string {
	extraResources := make(map[string][]string)

	extraResources["fleet-local"] = []string{"secrets"}
	return extraResources
}

// skipClusterObjects implements the ExcludeFilter and skips all cluster resources
// other than settings.harvesterhci.io
func (module harvesterModule) skipClusterObjects(gv schema.GroupVersion, resource v1.APIResource) bool {
	if gv.Group == "harvesterhci.io" && gv.Version == "v1beta1" && resource.Kind == "Setting" {
		logrus.Debugf("processing object %s with gv %s\n", resource.String(), gv.String())
		return false
	}
	logrus.Debugf("skipping object %s with gv %s\n", resource.String(), gv.String())
	return true
}

func (module harvesterModule) toClusterObj(b []byte, groupVersion, kind string, resources ...string) (interface{}, error) {
	jsonParsed, err := module.c.toObjCommon(b, groupVersion, kind)

	if err != nil {
		return nil, err
	}

	switch kind {
	case "Setting":
		currentItems, _ := jsonParsed.S("items").Data().([]interface{})
		logrus.Debugf("Whole items in cluster: %v", currentItems)
		var newItems []interface{}
		for _, item := range currentItems {
			gItem := gabs.Wrap(item)
			logrus.Debugf("processing setting %v", gItem.S("metadata", "name").Data().(string))
			if !slice.ContainsString(ignoreHarvesterSettingsList, gItem.S("metadata", "name").Data().(string)) {
				logrus.Debugf("Prepare to append item: %v", gItem.Data().(map[string]interface{}))
				newItems = append(newItems, item)
			}
		}
		if _, err := jsonParsed.Set(newItems, "items"); err != nil {
			return nil, err
		}
	default:
		// undefined resource, just logged it.
		logrus.Warnf("Could not handle kind %s", kind)
	}
	return jsonParsed.Data(), nil
}

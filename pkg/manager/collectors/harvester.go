package collectors

import (
	"path/filepath"

	"github.com/Jeffail/gabs/v2"
	"github.com/sirupsen/logrus"
)

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
			logrus.Error("Unable to fetch namespaced resources")
			return
		}

		for name, obj := range objs {
			file := filepath.Join(dir, name+".yaml")
			logrus.Debugf("Prepare to encode to yaml file path: %s", file)
			module.c.encodeFunc(obj, file, module.c.errorLog)
		}
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
			logrus.Debugf("whole items: %v", currentItems)
			var newItems []interface{}
			for _, item := range currentItems {
				gItem := gabs.Wrap(item)
				if find := gItem.S("type").Data().(string) == "rke.cattle.io/machine-plan"; find {
					logrus.Debugf("prepare to append item: %v", gItem.Data().(map[string]interface{}))
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

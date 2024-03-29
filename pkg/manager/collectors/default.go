package collectors

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type defaultModule struct {
	c          *common
	name       string
	nameSpaces []string
}

func NewDefaultModule(common *common, name string, ns []string) *defaultModule {
	return &defaultModule{
		c:          common,
		name:       name,
		nameSpaces: ns,
	}
}

func (module defaultModule) generateYAMLs() {
	logrus.Infof("[%s] generate YAMLs, yamlsDir: %s", module.name, module.c.yamlsDir)

	// Namespaced scope: all resources
	namespaces := []string{"default", "kube-system", "cattle-system"}
	namespaces = append(namespaces, module.nameSpaces...)

	done := make(map[string]struct{})
	for _, namespace := range namespaces {
		if _, ok := done[namespace]; ok {
			continue
		}

		namespacedDir := filepath.Join(module.c.yamlsDir, "namespaced", namespace)
		module.generateDiscoveredNamespacedYAMLs(namespace, namespacedDir, module.c.errorLog)

		done[namespace] = struct{}{}
	}

}

func (module defaultModule) toObj(b []byte, groupVersion, kind string, resources ...string) (interface{}, error) {
	jsonParsed, err := module.c.toObjCommon(b, groupVersion, kind)

	if err != nil {
		return nil, err
	}

	return jsonParsed.Data(), nil
}

func (module defaultModule) generateDiscoveredNamespacedYAMLs(namespace string, dir string, errLog io.Writer) {
	objs, err := module.c.discovery.ResourcesForNamespace(module.toObj, namespace, module.c.exclude, errLog)

	if err != nil {
		logrus.WithError(err).Error("Unable to fetch namespaced resources")
		fmt.Fprintf(module.c.errorLog, "Unable to fetch namespaced resources: %v\n", err)
		return
	}

	for name, obj := range objs {
		file := filepath.Join(dir, name+".yaml")
		module.c.encodeFunc(obj, file, errLog)
	}
}

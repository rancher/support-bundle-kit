package collectors

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type clusterModule struct {
	c    *common
	name string
}

func NewClusterModule(common *common, name string) *clusterModule {
	return &clusterModule{
		c:    common,
		name: name,
	}
}

func (module clusterModule) generateYAMLs() {
	logrus.Infof("[%s] generate YAMLs, yamlsDir: %s", module.name, module.c.yamlsDir)

	// Cluster scope
	globalDir := filepath.Join(module.c.yamlsDir, "cluster")
	objs, err := module.c.discovery.ResourcesForCluster(module.toObj, module.c.exclude, module.c.errorLog)

	if err != nil {
		logrus.WithError(err).Error("Unable to fetch cluster resources")
		fmt.Fprintf(module.c.errorLog, "Unable to fetch cluster resources: %v\n", err)
		return
	}

	for name, obj := range objs {
		file := filepath.Join(globalDir, name+".yaml")
		module.c.encodeFunc(obj, file, module.c.errorLog)
	}
}

func (module clusterModule) toObj(b []byte, groupVersion, kind string, resources ...string) (interface{}, error) {
	jsonParsed, err := module.c.toObjCommon(b, groupVersion, kind)

	if err != nil {
		return nil, err
	}

	return jsonParsed.Data(), nil
}

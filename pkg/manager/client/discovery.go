package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type DiscoveryClient struct {
	Context         context.Context
	discoveryClient *discovery.DiscoveryClient
}

func NewDiscoveryClient(ctx context.Context, config *rest.Config) (*DiscoveryClient, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return &DiscoveryClient{
		Context:         ctx,
		discoveryClient: discoveryClient,
	}, nil
}

func toObj(b []byte) interface{} {
	var objMap map[string]interface{}

	r := bytes.NewReader(b)
	d := json.NewDecoder(r)

	if err := d.Decode(&objMap); err != nil {
		logrus.Errorf("Unable to decode json.")
		return nil
	}

	// Check that the obj has items
	items, ok := objMap["items"].([]interface{})
	if ok {
		if len(items) > 0 {
			return objMap
		}
	}

	return nil
}

// Get all the namespaced resources for a given namespace
func (dc *DiscoveryClient) ResourcesForNamespace(namespace string) map[string]interface{} {
	objs := make(map[string]interface{})

	lists, err := dc.discoveryClient.ServerPreferredResources()
	if err != nil {
		return objs
	}

	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range list.APIResources {
			if !resource.Namespaced {
				continue
			}

			// I would like to build the URL with rest client
			// methods, but I was not able to.  It might be
			// possible if a new rest client is created each
			// time with the GroupVersion
			url := fmt.Sprintf("/apis/%s/namespaces/%s/%s", gv.String(), namespace, resource.Name)

			result := dc.discoveryClient.RESTClient().Get().AbsPath(url).Do(dc.Context)

			// It is likely that errors can occur.
			if result.Error() != nil {
				logrus.Tracef("Failed to get %s: %v", url, result.Error())
				continue
			}

			// This produces a byte array of json.
			b, err := result.Raw()

			if err == nil {
				obj := toObj(b)
				if obj != nil {
					objs[resource.Name] = obj
				}
			}
		}
	}

	return objs

}

// Get the cluster level resources
func (dc *DiscoveryClient) ResourcesForCluster() map[string]interface{} {
	objs := make(map[string]interface{})

	lists, err := dc.discoveryClient.ServerPreferredResources()
	if err != nil {
		return objs
	}

	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range list.APIResources {
			if resource.Namespaced {
				continue
			}

			url := fmt.Sprintf("/apis/%s/%s", gv.String(), resource.Name)

			result := dc.discoveryClient.RESTClient().Get().AbsPath(url).Do(dc.Context)

			// It is likely that errors can occur.
			if result.Error() != nil {
				logrus.Tracef("Failed to get %s: %v", url, result.Error())
				continue
			}

			b, err := result.Raw()

			if err == nil {
				obj := toObj(b)
				if obj != nil {
					objs[resource.Name] = obj
				} else {
					logrus.Tracef("%s is empty", url)
				}
			}
		}
	}

	return objs

}

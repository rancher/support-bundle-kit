package client

import (
	"context"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

const (
	discoveryBurst = 10000
	discoveryQPS   = 10000
)

type ParseResult func(b []byte, groupVersion, kind string, resources ...string) (interface{}, error)
type ExcludeFilter func(schema.GroupVersion, metav1.APIResource) bool

type DiscoveryClient struct {
	Context         context.Context
	discoveryClient *discovery.DiscoveryClient
}

func NewDiscoveryClient(ctx context.Context, config *rest.Config) (*DiscoveryClient, error) {
	newConfig := rest.CopyConfig(config)
	newConfig.Burst = discoveryBurst
	newConfig.QPS = discoveryQPS

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(newConfig)
	if err != nil {
		return nil, err
	}

	return &DiscoveryClient{
		Context:         ctx,
		discoveryClient: discoveryClient,
	}, nil
}

// Get extra resource/namespace and try to do specific filter with module name
func (dc *DiscoveryClient) SpecificResourcesForNamespace(toObj ParseResult, moduleName, namespace string, targetResource []string, errLog io.Writer) (map[string]interface{}, error) {

	// If we upgrade to golang v1.18, use slice.contain to replcae checing
	resourceChecking := make(map[string]bool)
	for _, resource := range targetResource {
		resourceChecking[resource] = true
	}

	objs := make(map[string]interface{})

	lists, err := dc.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
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

			if _, exists := resourceChecking[resource.Name]; !exists {
				continue
			}
			// I would like to build the URL with rest client
			// methods, but I was not able to.  It might be
			// possible if a new rest client is created each
			// time with the GroupVersion
			prefix := "apis"
			if gv.String() == "v1" {
				prefix = "api"
			}
			url := fmt.Sprintf("/%s/%s/namespaces/%s/%s", prefix, gv.String(), namespace, resource.Name)

			result := dc.discoveryClient.RESTClient().Get().AbsPath(url).Do(dc.Context)

			// It is likely that errors can occur.
			if result.Error() != nil {
				logrus.Tracef("Failed to get %s: %v", url, result.Error())
				fmt.Fprintf(errLog, "Failed to get %s: %v\n", url, result.Error())
				continue
			}

			// This produces a byte array of json.
			b, err := result.Raw()

			if err == nil {
				obj, err := toObj(b, gv.String(), resource.Kind, resource.Name)
				if err != nil {
					// This is unexpected. Log, but continue to try other resources.
					logrus.Errorf("Failed to parse objects received from %s: %v", url, result.Error())
					fmt.Fprintf(errLog, "Failed to parse objects received from %s: %v\n", url, result.Error())
					continue
				}
				// skip empty object, which will cause useless zero item yaml file
				if obj != nil {
					objs[gv.String()+"/"+resource.Name] = obj
				} else {
					logrus.Debugf("No %s/%s resource %s in namespace %s, skip", gv.String(), resource.Kind, resource.Name, namespace)
				}
			}

		}
	}

	return objs, nil
}

func (dc *DiscoveryClient) ResourcesForNamespace(toObj ParseResult, namespace string, exclude ExcludeFilter, errLog io.Writer) (map[string]interface{}, error) {
	objs := make(map[string]interface{})

	lists, err := dc.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
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

			if exclude(gv, resource) {
				continue
			}

			// I would like to build the URL with rest client
			// methods, but I was not able to.  It might be
			// possible if a new rest client is created each
			// time with the GroupVersion
			prefix := "apis"
			if gv.String() == "v1" {
				prefix = "api"
			}
			url := fmt.Sprintf("/%s/%s/namespaces/%s/%s", prefix, gv.String(), namespace, resource.Name)

			result := dc.discoveryClient.RESTClient().Get().AbsPath(url).Do(dc.Context)

			// It is likely that errors can occur.
			if result.Error() != nil {
				logrus.Tracef("Failed to get %s: %v", url, result.Error())
				fmt.Fprintf(errLog, "Failed to get %s: %v\n", url, result.Error())
				continue
			}

			// This produces a byte array of json.
			b, err := result.Raw()

			if err == nil {
				obj, err := toObj(b, gv.String(), resource.Kind)
				if err != nil {
					// This is unexpected. Log, but continue to try other resources.
					logrus.Errorf("Failed to parse objects received from %s: %v", url, result.Error())
					fmt.Fprintf(errLog, "Failed to parse objects received from %s: %v\n", url, result.Error())
					continue
				}
				// skip empty object, which will cause useless zero item yaml file
				if obj != nil {
					objs[gv.String()+"/"+resource.Name] = obj
				} else {
					logrus.Debugf("No %s/%s resource %s in namespace %s, skip", gv.String(), resource.Kind, resource.Name, namespace)
				}
			}
		}
	}

	return objs, nil
}

// Get the cluster level resources
func (dc *DiscoveryClient) ResourcesForCluster(toObj ParseResult, exclude ExcludeFilter, errLog io.Writer) (map[string]interface{}, error) {
	objs := make(map[string]interface{})

	lists, err := dc.discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
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

			if exclude(gv, resource) {
				continue
			}

			prefix := "apis"
			if gv.String() == "v1" {
				prefix = "api"
			}
			url := fmt.Sprintf("/%s/%s/%s", prefix, gv.String(), resource.Name)

			result := dc.discoveryClient.RESTClient().Get().AbsPath(url).Do(dc.Context)

			// It is likely that errors can occur.
			if result.Error() != nil {
				logrus.Tracef("Failed to get %s: %v", url, result.Error())
				fmt.Fprintf(errLog, "Failed to get %s: %v\n", url, result.Error())
				continue
			}

			b, err := result.Raw()

			if err == nil {
				obj, err := toObj(b, gv.String(), resource.Kind)
				if err != nil {
					// This is unexpected. Log, but continue to try other resources.
					logrus.Errorf("Failed to parse objects received from %s: %v", url, result.Error())
					fmt.Fprintf(errLog, "Failed to parse objects received from %s: %v\n", url, result.Error())
					continue
				}
				// skip empty object
				if obj != nil {
					objs[gv.String()+"/"+resource.Name] = obj
				}
			}
		}
	}

	return objs, nil
}

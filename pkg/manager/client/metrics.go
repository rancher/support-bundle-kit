package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type MetricsClient struct {
	Context   context.Context
	clientset *versioned.Clientset
}

func NewMetricsClient(ctx context.Context, config *rest.Config) (*MetricsClient, error) {
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &MetricsClient{
		Context:   ctx,
		clientset: clientset,
	}, nil
}

func (c *MetricsClient) GetAllNodeMetrics() (runtime.Object, error) {
	return c.clientset.MetricsV1beta1().NodeMetricses().List(c.Context, metav1.ListOptions{})
}

func (c *MetricsClient) GetAllPodMetrics(namespace string) (runtime.Object, error) {
	return c.clientset.MetricsV1beta1().PodMetricses(namespace).List(c.Context, metav1.ListOptions{})
}

package client

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesClient struct {
	Context   context.Context
	namespace string
	clientSet *kubernetes.Clientset
}

func NewKubernetesStore(ctx context.Context, namespace string, config *rest.Config) (*KubernetesClient, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &KubernetesClient{
		Context:   ctx,
		namespace: namespace,
		clientSet: clientSet,
	}, nil
}

func (k *KubernetesClient) GetNamespace() (*corev1.Namespace, error) {
	return k.clientSet.CoreV1().Namespaces().Get(k.Context, k.namespace, metav1.GetOptions{})
}

func (k *KubernetesClient) GetKubernetesVersion() (*version.Info, error) {
	return k.clientSet.Discovery().ServerVersion()
}

func (k *KubernetesClient) GetAllPodsList() (runtime.Object, error) {
	return k.clientSet.CoreV1().Pods(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetPodsListByLabels(labels string) (*corev1.PodList, error) {
	return k.clientSet.CoreV1().Pods(k.namespace).List(k.Context, metav1.ListOptions{LabelSelector: labels})
}

func (k *KubernetesClient) GetPodContainerLogRequest(podName, containerName string) *rest.Request {
	return k.clientSet.CoreV1().Pods(k.namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container:  containerName,
		Timestamps: true,
	})
}

func (k *KubernetesClient) GetAllServicesList() (runtime.Object, error) {
	return k.clientSet.CoreV1().Services(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllDeploymentsList() (runtime.Object, error) {
	return k.clientSet.AppsV1().Deployments(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetDeploymentsListByLabels(labels string) (*appsv1.DeploymentList, error) {
	return k.clientSet.AppsV1().Deployments(k.namespace).List(k.Context, metav1.ListOptions{LabelSelector: labels})
}

func (k *KubernetesClient) GetAllDaemonSetsList() (runtime.Object, error) {
	return k.clientSet.AppsV1().DaemonSets(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) CreateDaemonSets(daemonSet *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	return k.clientSet.AppsV1().DaemonSets(k.namespace).Create(k.Context, daemonSet, metav1.CreateOptions{})
}

func (k *KubernetesClient) DeleteDaemonSets(name string) error {
	return k.clientSet.AppsV1().DaemonSets(k.namespace).Delete(k.Context, name, metav1.DeleteOptions{})
}

func (k *KubernetesClient) GetAllStatefulSetsList() (runtime.Object, error) {
	return k.clientSet.AppsV1().StatefulSets(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllJobsList() (runtime.Object, error) {
	return k.clientSet.BatchV1().Jobs(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllCronJobsList() (runtime.Object, error) {
	return k.clientSet.BatchV1beta1().CronJobs(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllNodesList() (runtime.Object, error) {
	return k.clientSet.CoreV1().Nodes().List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetNodesListByLabels(labels string) (*corev1.NodeList, error) {
	return k.clientSet.CoreV1().Nodes().List(k.Context, metav1.ListOptions{LabelSelector: labels})
}

func (k *KubernetesClient) GetAllEventsList() (runtime.Object, error) {
	return k.clientSet.CoreV1().Events(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllConfigMaps() (runtime.Object, error) {
	return k.clientSet.CoreV1().ConfigMaps(k.namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllVolumeAttachments() (runtime.Object, error) {
	return k.clientSet.StorageV1().VolumeAttachments().List(k.Context, metav1.ListOptions{})
}

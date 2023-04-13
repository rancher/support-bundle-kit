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
	clientSet *kubernetes.Clientset
}

func NewKubernetesClient(ctx context.Context, config *rest.Config) (*KubernetesClient, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &KubernetesClient{
		Context:   ctx,
		clientSet: clientSet,
	}, nil
}

func (k *KubernetesClient) GetNamespace(namespace string) (*corev1.Namespace, error) {
	return k.clientSet.CoreV1().Namespaces().Get(k.Context, namespace, metav1.GetOptions{})
}

func (k *KubernetesClient) GetKubernetesVersion() (*version.Info, error) {
	return k.clientSet.Discovery().ServerVersion()
}

func (k *KubernetesClient) GetAllPodsList(namespace string) (runtime.Object, error) {
	return k.clientSet.CoreV1().Pods(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetPodsListByLabels(namespace string, labels string) (*corev1.PodList, error) {
	return k.clientSet.CoreV1().Pods(namespace).List(k.Context, metav1.ListOptions{LabelSelector: labels})
}

func (k *KubernetesClient) GetPodContainerLogRequest(namespace, podName, containerName string) *rest.Request {
	return k.clientSet.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container:  containerName,
		Timestamps: true,
	})
}

func (k *KubernetesClient) GetAllServicesList(namespace string) (runtime.Object, error) {
	return k.clientSet.CoreV1().Services(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllDeploymentsList(namespace string) (runtime.Object, error) {
	return k.clientSet.AppsV1().Deployments(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetDeploymentsListByLabels(namespace, labels string) (*appsv1.DeploymentList, error) {
	return k.clientSet.AppsV1().Deployments(namespace).List(k.Context, metav1.ListOptions{LabelSelector: labels})
}

func (k *KubernetesClient) GetAllDaemonSetsList(namespace string) (runtime.Object, error) {
	return k.clientSet.AppsV1().DaemonSets(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) CreateDaemonSets(namespace string, daemonSet *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	return k.clientSet.AppsV1().DaemonSets(namespace).Create(k.Context, daemonSet, metav1.CreateOptions{})
}

func (k *KubernetesClient) DeleteDaemonSets(namespace, name string) error {
	return k.clientSet.AppsV1().DaemonSets(namespace).Delete(k.Context, name, metav1.DeleteOptions{})
}

func (k *KubernetesClient) GetDaemonSetBy(namespace, name string) (*appsv1.DaemonSet, error) {
	return k.clientSet.AppsV1().DaemonSets(namespace).Get(k.Context, name, metav1.GetOptions{})
}

func (k *KubernetesClient) GetAllStatefulSetsList(namespace string) (runtime.Object, error) {
	return k.clientSet.AppsV1().StatefulSets(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllJobsList(namespace string) (runtime.Object, error) {
	return k.clientSet.BatchV1().Jobs(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllCronJobsList(namespace string) (runtime.Object, error) {
	return k.clientSet.BatchV1beta1().CronJobs(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetNodeBy(name string) (*corev1.Node, error) {
	return k.clientSet.CoreV1().Nodes().Get(k.Context, name, metav1.GetOptions{})
}

func (k *KubernetesClient) GetAllNodesList() (runtime.Object, error) {
	return k.clientSet.CoreV1().Nodes().List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetNodesListByLabels(labels string) (*corev1.NodeList, error) {
	return k.clientSet.CoreV1().Nodes().List(k.Context, metav1.ListOptions{LabelSelector: labels})
}

func (k *KubernetesClient) GetAllEventsList(namespace string) (runtime.Object, error) {
	return k.clientSet.CoreV1().Events(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllConfigMaps(namespace string) (runtime.Object, error) {
	return k.clientSet.CoreV1().ConfigMaps(namespace).List(k.Context, metav1.ListOptions{})
}

func (k *KubernetesClient) GetAllVolumeAttachments() (runtime.Object, error) {
	return k.clientSet.StorageV1().VolumeAttachments().List(k.Context, metav1.ListOptions{})
}

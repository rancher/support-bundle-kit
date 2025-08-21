package manager

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rancher/support-bundle-kit/pkg/types"
)

type AgentDaemonSet struct {
	sbm *SupportBundleManager
}

func (a *AgentDaemonSet) getDaemonSetName() string {
	return fmt.Sprintf("supportbundle-agent-%s", a.sbm.BundleName)
}

func (a *AgentDaemonSet) Create(image string, managerURL string) (*appsv1.DaemonSet, error) {
	dsName := a.getDaemonSetName()
	logrus.Debugf("Creating daemonset %s with image %s", dsName, image)

	// get manager pod for owner reference
	labels := fmt.Sprintf("app=%s,%s=%s", types.SupportBundleManager, types.SupportBundleLabelKey, a.sbm.BundleName)

	pods, err := a.sbm.k8s.GetPodsListByLabels(a.sbm.PodNamespace, labels)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, errors.New("no support bundle manager pod found")
	}

	if len(pods.Items) != 1 {
		return nil, errors.New("more than one support bundle manager pods are found")
	}
	managerPod := pods.Items[0]

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dsName,
			Namespace: a.sbm.PodNamespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					// not sure why managerPod has empty Kind and APIVersion
					Name:       managerPod.Name,
					Kind:       "Pod",
					UID:        managerPod.UID,
					APIVersion: "v1",
				},
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                       types.SupportBundleAgent,
					types.SupportBundleLabelKey: a.sbm.BundleName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                       types.SupportBundleAgent,
						types.SupportBundleLabelKey: a.sbm.BundleName,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: a.sbm.getNodeSelector(),
					Tolerations:  a.sbm.getTaintToleration(),
					Containers: []corev1.Container{
						{
							Name:            "agent",
							Image:           image,
							Args:            []string{"/usr/bin/support-bundle-collector.sh"},
							ImagePullPolicy: corev1.PullPolicy(a.sbm.ImagePullPolicy),
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{"SYSLOG"},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SUPPORT_BUNDLE_HOST_PATH",
									Value: "/host",
								},
								{
									Name: "SUPPORT_BUNDLE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name:  "SUPPORT_BUNDLE_MANAGER_URL",
									Value: managerURL,
								},
								{
									Name:  "SUPPORT_BUNDLE_COLLECTOR",
									Value: a.sbm.SpecifyCollector,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "host",
									MountPath: "/host",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "host",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/",
								},
							},
						},
					},
				},
			},
		},
	}

	if a.sbm.RegistrySecret != "" {
		daemonSet.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: a.sbm.RegistrySecret,
			},
		}
	}

	switch a.sbm.SpecifyCollector {
	case "longhorn":
		a.prepareDaemonSetForLonghorn(daemonSet)
	}

	return a.sbm.k8s.CreateDaemonSets(a.sbm.PodNamespace, daemonSet)
}

func (a *AgentDaemonSet) prepareDaemonSetForLonghorn(daemonset *appsv1.DaemonSet) {
	daemonset.Spec.Template.Spec.Containers[0].Env = append(daemonset.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  "LONGHORN_LOG_PATH",
		Value: os.Getenv("LONGHORN_LOG_PATH"),
	})
}

func (a *AgentDaemonSet) Cleanup() error {
	dsName := a.getDaemonSetName()
	err := a.sbm.k8s.DeleteDaemonSets(a.sbm.PodNamespace, dsName)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

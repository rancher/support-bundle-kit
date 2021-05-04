package client

import (
	"context"

	harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/harvester/harvester/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type HarvesterClient struct {
	context   context.Context
	namespace string
	clientset *versioned.Clientset
}

func NewHarvesterClient(ctx context.Context, namespace string, config *rest.Config) (*HarvesterClient, error) {
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &HarvesterClient{
		context:   ctx,
		namespace: namespace,
		clientset: clientset,
	}, nil
}

func (h *HarvesterClient) GetSupportBundle(name string) (*harvesterv1.SupportBundle, error) {
	return h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Get(h.context, name, metav1.GetOptions{})
}

func (h *HarvesterClient) GetSupportBundleState(name string) (string, error) {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return sb.Status.State, nil
}

func (h *HarvesterClient) SetSupportBundleError(name string, state string, errmsg string) error {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := sb.DeepCopy()
	toUpdate.Status.State = state

	harvesterv1.SupportBundleInitialized.False(toUpdate)
	harvesterv1.SupportBundleInitialized.Message(toUpdate, errmsg)

	_, err = h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Update(h.context, toUpdate, metav1.UpdateOptions{})
	return err
}

func (h *HarvesterClient) UpdateSupportBundleStatus(name string, state string) error {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := sb.DeepCopy()
	toUpdate.Status.State = state
	_, err = h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Update(h.context, toUpdate, metav1.UpdateOptions{})
	return err
}

func (h *HarvesterClient) UpdateSupportBundleStatus2(name string, state string, filename string, filesize int64) error {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := sb.DeepCopy()
	toUpdate.Status.State = state
	toUpdate.Status.Filename = filename
	toUpdate.Status.Filesize = filesize
	if state == "ready" {
		harvesterv1.SupportBundleInitialized.True(toUpdate)
	}
	_, err = h.clientset.HarvesterhciV1beta1().SupportBundles(h.namespace).Update(h.context, toUpdate, metav1.UpdateOptions{})
	return err
}

func (h *HarvesterClient) GetAllKeypairs() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().KeyPairs(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllPreferences() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Preferences(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllSettings() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Settings().List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllUpgrades() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Upgrades(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllUsers() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Users().List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineBackups() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineBackups(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineBackupContents() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineBackupContents(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineImages() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineImages(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineRestores() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineRestores(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineTemplates() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineTemplates(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineTemplateVersions() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineTemplateVersions(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachines() (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachines(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineInstances() (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachineInstances(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineInstanceMigrations() (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachineInstanceMigrations(h.namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllDataVolumes() (runtime.Object, error) {
	return h.clientset.CdiV1beta1().DataVolumes(h.namespace).List(h.context, metav1.ListOptions{})
}

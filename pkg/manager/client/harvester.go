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
	clientset *versioned.Clientset
}

func NewHarvesterClient(ctx context.Context, config *rest.Config) (*HarvesterClient, error) {
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &HarvesterClient{
		context:   ctx,
		clientset: clientset,
	}, nil
}

func (h *HarvesterClient) GetSupportBundle(namespace, name string) (*harvesterv1.SupportBundle, error) {
	return h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Get(h.context, name, metav1.GetOptions{})
}

func (h *HarvesterClient) GetSupportBundleState(namespace, name string) (string, error) {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return sb.Status.State, nil
}

func (h *HarvesterClient) SetSupportBundleError(namespace string, name string, state string, errmsg string) error {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := sb.DeepCopy()
	toUpdate.Status.State = state

	harvesterv1.SupportBundleInitialized.False(toUpdate)
	harvesterv1.SupportBundleInitialized.Message(toUpdate, errmsg)

	_, err = h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Update(h.context, toUpdate, metav1.UpdateOptions{})
	return err
}

func (h *HarvesterClient) UpdateSupportBundleStatus(namespace string, name string, state string) error {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Get(h.context, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := sb.DeepCopy()
	toUpdate.Status.State = state
	_, err = h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Update(h.context, toUpdate, metav1.UpdateOptions{})
	return err
}

func (h *HarvesterClient) UpdateSupportBundleStatus2(namespace string, name string, state string, filename string, filesize int64) error {
	sb, err := h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Get(h.context, name, metav1.GetOptions{})
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
	_, err = h.clientset.HarvesterhciV1beta1().SupportBundles(namespace).Update(h.context, toUpdate, metav1.UpdateOptions{})
	return err
}

func (h *HarvesterClient) GetAllKeypairs(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().KeyPairs(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllPreferences(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Preferences(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllSettings() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Settings().List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllUpgrades(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Upgrades(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllUsers() (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().Users().List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineBackups(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineBackups(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineBackupContents(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineBackupContents(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineImages(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineImages(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineRestores(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineRestores(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineTemplates(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineTemplates(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineTemplateVersions(namespace string) (runtime.Object, error) {
	return h.clientset.HarvesterhciV1beta1().VirtualMachineTemplateVersions(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachines(namespace string) (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachines(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineInstances(namespace string) (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachineInstances(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllVirtualMachineInstanceMigrations(namespace string) (runtime.Object, error) {
	return h.clientset.KubevirtV1().VirtualMachineInstanceMigrations(namespace).List(h.context, metav1.ListOptions{})
}

func (h *HarvesterClient) GetAllDataVolumes(namespace string) (runtime.Object, error) {
	return h.clientset.CdiV1beta1().DataVolumes(namespace).List(h.context, metav1.ListOptions{})
}

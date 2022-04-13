package integration

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher/support-bundle-kit/pkg/simulator/objects"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

var _ = Describe("Process Support Bundle", func() {
	var o *objects.ObjectManager
	var err error
	BeforeEach(func() {
		Eventually(func() error {
			o, err = objects.NewObjectManager(ctx, a.Config, samplesPath)
			if err != nil {
				return err
			}
			return o.WaitForNamespaces(15 * time.Second)
		}, 5, 60).ShouldNot(HaveOccurred())
	})

	It("Load cluster scoped objects", func() {
		Eventually(func() error {
			return o.CreateUnstructuredClusterObjects()
		}, 5, 60).ShouldNot(HaveOccurred())
	})

	It("Load namespace scoped objects", func() {
		Eventually(func() error {
			return o.CreateUnstructuredObjects()
		}, 5, 60).ShouldNot(HaveOccurred())
	})

	It("Verify Pods", func() {
		By("Verify pod objects")
		{
			Eventually(func() error {
				objs, err := objects.GenerateObjects(samplePodSpec)
				if err != nil {
					return err
				}
				for _, obj := range objs {
					_, err := o.FetchObject(obj)
					if err != nil {
						return err
					}
				}
				return nil
			}, 5, 60).ShouldNot(HaveOccurred())

		}
	})

	It("Verify Nodes", func() {
		By("Verify node addresses are localhost")
		{
			Eventually(func() error {
				kc, err := kubernetes.NewForConfig(a.Config)
				if err != nil {
					return err
				}

				nodes, err := kc.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
				if err != nil {
					return err
				}

				for _, node := range nodes.Items {
					nodeObj, err := kc.CoreV1().Nodes().Get(ctx, node.GetName(), metav1.GetOptions{})
					if err != nil {
						return err
					}

					for _, address := range nodeObj.Status.Addresses {
						if address.Address != "localhost" {
							return fmt.Errorf("expect addresses to be localhost but found %s for node %s", address.Address, nodeObj.GetName())
						}
					}
				}
				return nil
			}, 5, 60).ShouldNot(HaveOccurred())
		}
	})
})

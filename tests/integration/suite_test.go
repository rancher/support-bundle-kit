package integration

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher/support-bundle-kit/pkg/simulator/apiserver"
	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
	"github.com/rancher/support-bundle-kit/pkg/simulator/kubelet"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	setupTimeout  = 600
	samplesPath   = "./sampleSupportBundle"
	samplePodSpec = "./sampleSupportBundle/yamls/namespaced/harvester-system/v1/pods.yaml"
)

func TestSim(t *testing.T) {
	defer GinkgoRecover()
	RegisterFailHandler(Fail)
	RunSpecs(t, "sim integration")
}

var (
	ctx    context.Context
	cancel context.CancelFunc
	a      apiserver.APIServerConfig
	dir    string
	eg     *errgroup.Group
	egctx  context.Context
)

var _ = BeforeSuite(func(done Done) {
	defer close(done)
	var err error
	By("starting test cluster")
	ctx, cancel = context.WithCancel(context.TODO())

	dir, err = ioutil.TempDir("/tmp", "integration-")
	Expect(err).ToNot(HaveOccurred())

	certificates, err := certs.GenerateCerts([]string{"localhost", "127.0.0.1"},
		dir)
	Expect(err).ToNot(HaveOccurred())
	a.Certs = certificates

	etcd, err := etcd.RunEmbeddedEtcd(ctx, dir, certificates)
	Expect(err).ToNot(HaveOccurred())
	a.Etcd = etcd

	Expect(a.GenerateKubeConfig(filepath.Join(dir, "admin.kubeconfig"))).ToNot(HaveOccurred())

	eg, egctx = errgroup.WithContext(ctx)
	Expect(err).ToNot(HaveOccurred())

	eg.Go(func() error {
		return a.RunAPIServer(egctx)
	})

	// run fake kubelet
	k, err := kubelet.NewKubeletSimulator(ctx, certificates, samplesPath)
	Expect(err).ToNot(HaveOccurred())
	eg.Go(func() error {
		return k.RunFakeKubelet()
	})

	eg.Go(func() error {
		return eg.Wait()
	})
}, setupTimeout)

var _ = AfterSuite(func(done Done) {
	defer os.RemoveAll(dir)
	defer close(done)
	cancel()
}, setupTimeout)

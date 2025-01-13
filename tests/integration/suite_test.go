package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/rancher/support-bundle-kit/pkg/simulator/apiserver"
	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
	"github.com/rancher/support-bundle-kit/pkg/simulator/kubelet"
	"github.com/rancher/support-bundle-kit/pkg/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	setupTimeout = 600

	sampleBundleZip = "./sampleSupportBundle.zip"
)

var samplesPath, samplePodSpec string

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

	ctx, cancel = context.WithCancel(context.TODO())

	dir, err = os.MkdirTemp("/tmp", "integration-")
	Expect(err).ToNot(HaveOccurred())

	// unzip support bundle contents
	By("extracting support bundle to temp directory")
	err = utils.UnzipSupportBundle(sampleBundleZip, dir)
	Expect(err).ToNot(HaveOccurred())
	samplesPath = filepath.Join(dir, "sampleSupportBundle")
	samplePodSpec = filepath.Join(samplesPath, "/yamls/namespaced/harvester-system/v1/pods.yaml")

	By("setting up test cluster")
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
		return a.RunAPIServer(egctx, apiserver.DefaultServiceClusterIP)
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
	close(done)
	os.RemoveAll(dir)
}, setupTimeout)

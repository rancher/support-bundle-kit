package integration

import (
	"archive/zip"
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rancher/support-bundle-kit/pkg/simulator/apiserver"
	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
	"github.com/rancher/support-bundle-kit/pkg/simulator/kubelet"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	setupTimeout = 600

	sampleBundleZip = "./sampleSupportBundle.zip"
)

var samplesRoot, samplesPath, samplePodSpec string

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

	By("extracting support bundle to temp directory")
	samplesRoot, err = unzipSupportBundle(sampleBundleZip)
	Expect(err).ToNot(HaveOccurred())

	samplesPath = filepath.Join(samplesRoot, "sampleSupportBundle")
	samplePodSpec = filepath.Join(samplesPath, "/yamls/namespaced/harvester-system/v1/pods.yaml")

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
	os.RemoveAll(dir)
	os.RemoveAll(samplesRoot)
	cancel()
}, setupTimeout)

func unzipSupportBundle(bundleZipFile string) (dest string, err error) {
	dest, err = ioutil.TempDir("/tmp", "support-bundle")
	if err != nil {
		return dest, err
	}

	r, err := zip.OpenReader(bundleZipFile)
	if err != nil {
		return dest, err
	}

	for _, f := range r.File {
		destPath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return dest, fmt.Errorf("invalid dest path %s", destPath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return dest, err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return dest, err
			}

			destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_CREATE, f.Mode())
			if err != nil {
				return dest, err
			}

			zFile, err := f.Open()
			if err != nil {
				return dest, err
			}

			if _, err = io.Copy(destFile, zFile); err != nil {
				return dest, err
			}
			zFile.Close()
			destFile.Close()
		}

	}
	return dest, nil
}

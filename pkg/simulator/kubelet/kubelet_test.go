package kubelet

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/utils"
)

const (
	bundleZipPath = "../../../tests/integration/sampleSupportBundle.zip"
)

var bundlePath string

func TestKubeletSimulator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp", "kubelet-")
	if err != nil {
		t.Fatalf("error creating a temp directory for kubelet: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.UnzipSupportBundle(bundleZipPath, tmpDir)
	if err != nil {
		t.Fatalf("Error during unzip operation %v", err)
	}
	certificates, err := certs.GenerateCerts([]string{"localhost", "127.0.0.1"},
		tmpDir)

	eg, egctx := errgroup.WithContext(context.Background())

	if err != nil {
		t.Fatalf("error generating certificates: %v", err)
	}

	bundlePath = filepath.Join(tmpDir, "sampleSupportBundle")
	k, err := NewKubeletSimulator(egctx, certificates, bundlePath)
	if err != nil {
		t.Fatalf("error creating new kubelet simulator")
	}

	eg.Go(func() error {
		return k.RunFakeKubelet()
	})

	eg.Go(func() error {
		return eg.Wait()
	})

	// run tests to fetch logs
	time.Sleep(10 * time.Second)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := http.Client{Transport: tr, Timeout: 10 * time.Second}
	resp, err := client.Get("https://localhost:10250/containerLogs/longhorn-system/backing-image-manager-d7ad-3cf5/backing-image-manager")
	if err != nil {
		t.Fatalf("error fetching logs from kubelet: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected to get status code 200 while reading container log but got %d", resp.StatusCode)
	}

	// read log from zip file as well
	resp, err = client.Get("https://localhost:10250/containerLogs/support-bundle-node-info/harvester-jb9lj/rke2-server")
	if err != nil {
		t.Fatalf("error fetching logs from kubelet: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected to get status code 200 while reading node logs from zip file but got %d", resp.StatusCode)
	}
}

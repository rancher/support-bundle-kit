package kubelet

import (
	"context"
	"crypto/tls"
	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	bundlePath = "../../../tests/integration/sampleSupportBundle"
)

func TestKubeletSimulator(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "kubelet-")
	if err != nil {
		t.Fatalf("error creating a temp directory for kubelet: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certificates, err := certs.GenerateCerts([]string{"localhost", "127.0.0.1"},
		tmpDir)

	eg, egctx := errgroup.WithContext(context.Background())

	if err != nil {
		t.Fatalf("error generating certificates: %v", err)
	}

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
	resp, err := client.Get("https://localhost:10250/containerLogs/longhorn-system/backing-image-manager-c00e-a3c7/backing-image-manager")
	if err != nil {
		t.Fatalf("error fetching logs from kubelet: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected to get status code 200 but got %d", resp.StatusCode)
	}

}

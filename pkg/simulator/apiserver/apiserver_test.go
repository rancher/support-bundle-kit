package apiserver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
)

func TestRunAPIServer(t *testing.T) {

	dir, err := os.MkdirTemp("/tmp", "apiserver-")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("error setting up temp directory for apiserver %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	a := NewAPIServerConfig(DefaultClientQPS, DefaultClientBurst)

	generatedCerts, err := certs.GenerateCerts([]string{"localhost", "127.0.0.1"}, dir)
	if err != nil {
		t.Fatalf("error generating certificates for sim %v", err)
	}
	a.Certs = generatedCerts

	etcdConfig, err := etcd.RunEmbeddedEtcd(context.TODO(), filepath.Join(dir), a.Certs)
	if err != nil {
		t.Fatalf("error setting up embedded etcdserver %v", err)
	}
	a.Etcd = etcdConfig

	err = a.GenerateKubeConfig(filepath.Join(dir, "admin.kubeconfig"))
	if err != nil {
		t.Fatalf("error generating kubeconfig %v", err)
	}

	err = a.RunAPIServer(ctx, DefaultServiceClusterIP)
	if err != nil {
		t.Fatalf("error running API Server %v", err)
	}
}

package crd

import (
	"context"
	"github.com/rancher/support-bundle-kit/pkg/simulator/apiserver"
	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteFiles(t *testing.T) {
	tmpFile, err := ioutil.TempFile("/tmp", "crd-")
	defer os.Remove(tmpFile.Name())
	if err != nil {
		t.Fatalf("error creating tmp file for crds: %v", err)
	}

	err = WriteFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("error writing crd spec: %v", err)
	}
}

func TestInstallCRD(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "apiserver-")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("error setting up temp directory for apiserver %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	a := apiserver.APIServerConfig{}

	generatedCerts, err := certs.GenerateCerts([]string{"localhost", "127.0.0.1"}, dir)
	if err != nil {
		t.Fatalf("error generating certificates for sim %v", err)
	}
	a.Certs = generatedCerts

	etcdConfig, err := etcd.RunEmbeddedEtcd(context.TODO(), filepath.Join(dir), a.Certs)
	if err != nil {
		t.Fatalf("error setting up embedded etcdserver")
	}
	a.Etcd = etcdConfig

	err = a.GenerateKubeConfig(filepath.Join(dir, "admin.kubeconfig"))
	if err != nil {
		t.Fatalf("error generating kubeconfig %v", err)
	}

	eg, egctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return a.RunAPIServer(egctx)
	})

	eg.Go(func() error {
		if err := eg.Wait(); err != nil {
			egctx.Done()
		}
		return nil
	})

	defer egctx.Done()
	time.Sleep(10 * time.Second)
	err = Create(egctx, a.Config)
	if err != nil {
		t.Fatalf("error installing crds :%v", err)
	}
}

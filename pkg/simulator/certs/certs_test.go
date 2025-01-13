package certs

import (
	"crypto/tls"
	"os"
	"testing"
)

func TestGenerateCerts(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "cert-unit-test")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatal("error creating certificate in /tmp")
	}

	c, err := GenerateCerts([]string{"localhost"}, dir)
	if err != nil {
		t.Fatal("error generating certificates")
	}

	_, err = tls.LoadX509KeyPair(c.CACert, c.CACertKey)
	if err != nil {
		t.Fatalf("error verifying CA keypair %v\n", err)
	}

	_, err = tls.LoadX509KeyPair(c.EtcdPeerCert, c.EtcdPeerCertKey)
	if err != nil {
		t.Fatalf("error verifying ETCD Peer keypair %v\n", err)
	}

	_, err = tls.LoadX509KeyPair(c.EtcdClientCert, c.EtcdClientCertKey)
	if err != nil {
		t.Fatalf("error verifying ETCD Client keypair %v\n", err)
	}

	_, err = tls.LoadX509KeyPair(c.APICert, c.APICertKey)
	if err != nil {
		t.Fatalf("error verifying APIServer keypair %v\n", err)
	}

	_, err = tls.LoadX509KeyPair(c.AdminCert, c.AdminCertKey)
	if err != nil {
		t.Fatalf("error verifying Admin keypair %v\n", err)
	}

	_, err = tls.LoadX509KeyPair(c.ServiceAccountCert, c.ServiceAccountCertKey)
	if err != nil {
		t.Fatalf("error verifying SA keypair %v\n", err)
	}

	_, err = tls.LoadX509KeyPair(c.KubeletCert, c.KubeletCertKey)
	if err != nil {
		t.Fatalf("error verifying Kubelet keypair %v\n", err)
	}
}

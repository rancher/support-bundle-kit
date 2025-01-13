package certs

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type CertInfo struct {
	Dir                   string
	CACert                string
	CACertKey             string
	EtcdPeerCert          string
	EtcdPeerCertKey       string
	EtcdClientCert        string
	EtcdClientCertKey     string
	APICert               string
	APICertKey            string
	KubeletCert           string
	KubeletCertKey        string
	AdminCert             string
	AdminCertKey          string
	ServiceAccountCert    string
	ServiceAccountCertKey string
}

func GenerateCerts(hosts []string, dir string) (*CertInfo, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	c := &CertInfo{}
	c.Dir = dir
	if err != nil {
		return nil, err
	}

	caTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * (24 * time.Hour)),

		Subject: pkix.Name{
			Organization: []string{"kubernetes"},
		},

		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	etcdServerTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(0).Add(serialNumber, big.NewInt(1)),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * (24 * time.Hour)),

		Subject: pkix.Name{
			Organization: []string{"kubernetes"},
		},

		SubjectKeyId:          []byte{1, 2, 3, 4, 6},
		DNSNames:              hosts,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	etcdClientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(0).Add(serialNumber, big.NewInt(2)),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * (24 * time.Hour)),

		Subject: pkix.Name{
			Organization: []string{"kubernetes"},
			CommonName:   "etcd-client",
		},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	apiServerTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(0).Add(serialNumber, big.NewInt(2)),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * (24 * time.Hour)),

		Subject: pkix.Name{
			Organization: []string{"kubernetes"},
			CommonName:   "kubernetes",
		},
		DNSNames:              hosts,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	adminTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(0).Add(serialNumber, big.NewInt(2)),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * (24 * time.Hour)),

		Subject: pkix.Name{
			Organization: []string{"system:masters"},
			CommonName:   "admin",
		},

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              hosts,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	saTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(0).Add(serialNumber, big.NewInt(2)),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * (24 * time.Hour)),

		Subject: pkix.Name{
			Organization: []string{"kubernetes"},
			CommonName:   "service-accounts",
		},

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	kubeletTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(0).Add(serialNumber, big.NewInt(2)),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * (24 * time.Hour)),

		Subject: pkix.Name{
			Organization: []string{"system:nodes"},
			CommonName:   "system:nodes:virtual-kubelet",
		},
		DNSNames:              hosts,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	caKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	serverKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	adminTemplateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	saTemplateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	apiServerKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	kubeletKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// CA keypair
	if c.CACertKey, err = ecPrivateKeyToFile(caKey, filepath.Join(dir, "ca", "ca-key.pem")); err != nil {
		return nil, err
	}
	if c.CACert, err = certToFile(caTemplate, caTemplate, &caKey.PublicKey, caKey, filepath.Join(dir, "ca", "ca-cert.pem")); err != nil {
		return nil, err
	}

	// ETCD peer keypair
	if c.EtcdPeerCertKey, err = ecPrivateKeyToFile(serverKey, filepath.Join(dir, "peer", "etcd-peer-key.pem")); err != nil {
		return nil, err
	}
	if c.EtcdPeerCert, err = certToFile(etcdServerTemplate, caTemplate, &serverKey.PublicKey, caKey, filepath.Join(dir, "peer", "etcd-peer-cert.pem")); err != nil {
		return nil, err
	}

	// ETCD client keypair
	if c.EtcdClientCertKey, err = ecPrivateKeyToFile(clientKey, filepath.Join(dir, "client", "etcd-client-key.pem")); err != nil {
		return nil, err
	}
	if c.EtcdClientCert, err = certToFile(etcdClientTemplate, caTemplate, &clientKey.PublicKey, caKey, filepath.Join(dir, "client", "etcd-client-cert.pem")); err != nil {
		return nil, err
	}

	// Admin keypair
	if c.AdminCertKey, err = ecPrivateKeyToFile(adminTemplateKey, filepath.Join(dir, "kubernetes", "admin-key.pem")); err != nil {
		return nil, err
	}
	if c.AdminCert, err = certToFile(adminTemplate, caTemplate, &adminTemplateKey.PublicKey, caKey, filepath.Join(dir, "kubernetes", "admin-cert.pem")); err != nil {
		return nil, err
	}

	// SA keypair
	if c.ServiceAccountCertKey, err = ecPrivateKeyToFile(saTemplateKey, filepath.Join(dir, "kubernetes", "sa-key.pem")); err != nil {
		return nil, err
	}
	if c.ServiceAccountCert, err = certToFile(saTemplate, caTemplate, &saTemplateKey.PublicKey, caKey, filepath.Join(dir, "kubernetes", "sa-cert.pem")); err != nil {
		return nil, err
	}

	// APIServer keypair
	if c.APICertKey, err = ecPrivateKeyToFile(apiServerKey, filepath.Join(dir, "kubernetes", "apiserver-key.pem")); err != nil {
		return nil, err
	}
	if c.APICert, err = certToFile(apiServerTemplate, caTemplate, &apiServerKey.PublicKey, caKey, filepath.Join(dir, "kubernetes", "apiserver-cert.pem")); err != nil {
		return nil, err
	}

	// Kubelet keypair
	if c.KubeletCertKey, err = ecPrivateKeyToFile(kubeletKey, filepath.Join(dir, "kubernetes", "kubelet-key.pem")); err != nil {
		return nil, err
	}
	if c.KubeletCert, err = certToFile(kubeletTemplate, caTemplate, &kubeletKey.PublicKey, caKey, filepath.Join(dir, "kubernetes", "kubelet-cert.pem")); err != nil {
		return nil, err
	}
	return c, nil
}

func certToFile(template *x509.Certificate, parent *x509.Certificate, publicKey *ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, path string) (string, error) {
	b, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, privateKey)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	if err := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: b}); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, buf.Bytes(), 0600)
}

func ecPrivateKeyToFile(key *ecdsa.PrivateKey, path string) (string, error) {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := pem.Encode(buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, buf.Bytes(), 0600)
}

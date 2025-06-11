package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/server/v3/embed"
)

type EtcdConfig struct {
	Endpoints []string
	TLS       *tls.Config
}

const (
	PeerPort   = "2380"
	ClientPort = "2379"
)

func RunEmbeddedEtcd(ctx context.Context, path string, certs *certs.CertInfo) (*EtcdConfig, error) {
	logrus.Info("Creating embedded etcd server")
	cfg := embed.NewConfig()

	cfg.Logger = "zap"
	cfg.LogLevel = "warn"

	cfg.Dir = filepath.Join(path, "embedded")
	cfg.AuthToken = ""

	scheme := "http"
	if certs != nil {
		scheme = "https"
	}
	cfg.ListenPeerUrls = []url.URL{{Scheme: scheme, Host: "localhost:" + PeerPort}}
	cfg.AdvertisePeerUrls = []url.URL{{Scheme: scheme, Host: "localhost:" + PeerPort}}
	cfg.ListenClientUrls = []url.URL{{Scheme: scheme, Host: "localhost:" + ClientPort}}
	cfg.AdvertiseClientUrls = []url.URL{{Scheme: scheme, Host: "localhost:" + ClientPort}}

	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)

	if err := os.MkdirAll(cfg.Dir, 0700); err != nil {
		return nil, err
	}

	if certs != nil {
		cfg.PeerTLSInfo.ServerName = "localhost"
		cfg.PeerTLSInfo.CertFile = certs.EtcdPeerCert
		cfg.PeerTLSInfo.KeyFile = certs.EtcdPeerCertKey
		cfg.PeerTLSInfo.TrustedCAFile = certs.CACert
		cfg.PeerTLSInfo.ClientCertAuth = true

		cfg.ClientTLSInfo.TrustedCAFile = certs.CACert
		cfg.ClientTLSInfo.ServerName = "localhost"
		cfg.ClientTLSInfo.CertFile = certs.EtcdPeerCert
		cfg.ClientTLSInfo.KeyFile = certs.EtcdPeerCertKey
		cfg.ClientTLSInfo.TrustedCAFile = certs.CACert
		cfg.ClientTLSInfo.ClientCertAuth = true
	}

	if enableUnsafeEtcdDisableFsyncHack, _ := strconv.ParseBool(os.Getenv("UNSAFE_E2E_HACK_DISABLE_ETCD_FSYNC")); enableUnsafeEtcdDisableFsyncHack {
		cfg.UnsafeNoFsync = true
	}

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}
	// Shutdown when context is closed
	go func() {
		<-ctx.Done()
		e.Close()
	}()

	clientConfig, err := cfg.ClientTLSInfo.ClientConfig()
	if err != nil {
		return nil, err
	}

	select {
	case <-e.Server.ReadyNotify():
		return &EtcdConfig{
			Endpoints: []string{cfg.AdvertiseClientUrls[0].String()},
			TLS:       clientConfig,
		}, nil
	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		return nil, fmt.Errorf("server took too long to start")
	case e := <-e.Err():
		return nil, e
	}
}

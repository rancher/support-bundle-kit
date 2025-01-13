package apiserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"

	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
)

const (
	DefaultServiceClusterIP = "10.53.0.1"
	DefaultClientQPS        = 100
	DefaultClientBurst      = 100
)

type APIServerConfig struct {
	Certs      *certs.CertInfo
	Etcd       *etcd.EtcdConfig
	KubeConfig string
	Config     *rest.Config
	QPS        float32
	Burst      int
}

const (
	APIVersionsSupported = "v1=true,api/beta=true,api/alpha=false"
)

func NewAPIServerConfig(qps float32, burst int) APIServerConfig {
	return APIServerConfig{
		QPS:   qps,
		Burst: burst,
	}
}

// RunAPIServer will bootstrap an API server with only core resources enabled
// No additional controllers will be scheduled
func (a *APIServerConfig) RunAPIServer(ctx context.Context, serviceClusterIP string) error {
	var err error
	s := options.NewServerRunOptions()
	s.Etcd.StorageConfig.Transport.ServerList = a.Etcd.Endpoints
	s.Etcd.StorageConfig.Transport.KeyFile = a.Certs.EtcdPeerCertKey
	s.Etcd.StorageConfig.Transport.CertFile = a.Certs.EtcdPeerCert
	s.Etcd.StorageConfig.Transport.TrustedCAFile = a.Certs.CACert

	s.APIEnablement.RuntimeConfig = map[string]string{
		"v1":       "true",
		"api/beta": "true",
	}
	s.AllowPrivileged = true
	s.ServiceAccountSigningKeyFile = a.Certs.ServiceAccountCertKey
	s.ServiceClusterIPRanges = serviceClusterIP + "/16"
	s.SecureServing.SecureServingOptions.ServerCert.CertKey.CertFile = a.Certs.APICert
	s.SecureServing.SecureServingOptions.ServerCert.CertKey.KeyFile = a.Certs.APICertKey
	s.SecureServing.SecureServingOptions.BindAddress = net.ParseIP("127.0.0.1")
	s.SecureServing.SecureServingOptions.BindPort = 6443
	s.Authentication.ServiceAccounts.KeyFiles = []string{a.Certs.ServiceAccountCertKey}
	s.Authentication.ServiceAccounts.Issuers = []string{"https://localhost:6443"}
	s.Authentication.ClientCert.ClientCA = a.Certs.CACert
	s.EventTTL = time.Duration(90 * 24 * time.Hour) // 90 days

	completedOptions, err := s.Complete()
	if err != nil {
		return err
	}

	if errs := completedOptions.Validate(); len(errs) != 0 {
		return utilerrors.NewAggregate(errs)
	}
	// Shutdown when context is closed
	go func() {
		<-ctx.Done()
		genericapiserver.RequestShutdown()
	}()

	return app.Run(completedOptions, genericapiserver.SetupSignalHandler())
}

// GenerateKubeConfig will generate KubeConfig to allow access to cluster
func (a *APIServerConfig) GenerateKubeConfig(path string) error {

	caCertByte, err := os.ReadFile(a.Certs.CACert)
	if err != nil {
		return fmt.Errorf("error read ca cert %v", err)
	}

	adminCertByte, err := os.ReadFile(a.Certs.AdminCert)
	if err != nil {
		return fmt.Errorf("error read admin cert %v", err)
	}

	adminCertKeyByte, err := os.ReadFile(a.Certs.AdminCertKey)
	if err != nil {
		return fmt.Errorf("error read admin cert key %v", err)
	}

	kc := kubeconfig.NewConfig()

	cluster := kubeconfig.NewCluster()
	cluster.CertificateAuthorityData = caCertByte
	cluster.Server = "https://localhost:6443"

	authInfo := kubeconfig.NewAuthInfo()
	authInfo.ClientCertificateData = adminCertByte
	authInfo.ClientKeyData = adminCertKeyByte

	kcContext := kubeconfig.NewContext()
	kcContext.AuthInfo = "default"
	kcContext.Cluster = "default"

	kc.Clusters["default"] = cluster
	kc.AuthInfos["default"] = authInfo
	kc.Contexts["default"] = kcContext
	kc.CurrentContext = "default"

	err = clientcmd.WriteToFile(*kc, path)
	if err != nil {
		return err
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*kc, nil)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}

	config.QPS = a.QPS
	config.Burst = a.Burst
	logrus.Infof("Client will be configured with QPS: %f, Burst: %d", config.QPS, config.Burst)

	a.Config = config

	return nil
}

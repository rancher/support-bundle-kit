package apiserver

import (
	"context"
	"fmt"
	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
	"io/ioutil"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/cmd/kube-apiserver/app"
	"strings"
)

type APIServerConfig struct {
	Certs      *certs.CertInfo
	Etcd       *etcd.EtcdConfig
	KubeConfig string
	Config     *rest.Config
}

const (
	APIVersionsSupported = "v1=true,api/beta=true,api/alpha=false"
)

// RunAPIServer will bootstrap an API server with only core resources enabled
// No additional controllers will be scheduled
func (a *APIServerConfig) RunAPIServer(ctx context.Context) error {
	var err error
	apiServer := app.NewAPIServerCommand()

	//set flag values
	if err = apiServer.Flags().Set("tls-cert-file", a.Certs.APICert); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("tls-private-key-file", a.Certs.APICert); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("client-ca-file", a.Certs.CACert); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("service-account-key-file", a.Certs.ServiceAccountCert); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("service-account-signing-key-file", a.Certs.ServiceAccountCertKey); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("service-account-issuer", "https://localhost:6443"); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("tls-cert-file", a.Certs.APICert); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("tls-private-key-file", a.Certs.APICertKey); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("runtime-config", APIVersionsSupported); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("enable-priority-and-fairness", "false"); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("service-cluster-ip-range", "10.53.0.0/16"); err != nil {
		return err
	}
	if err = apiServer.Flags().Set("allow-privileged", "true"); err != nil {
		return err
	}

	etcdList := strings.Join(a.Etcd.Endpoints, ",")
	if err = apiServer.Flags().Set("etcd-servers", etcdList); err != nil {
		return err
	}

	if a.Etcd.TLS != nil {
		if err = apiServer.Flags().Set("etcd-cafile", a.Certs.CACert); err != nil {
			return err
		}
		if err = apiServer.Flags().Set("etcd-certfile", a.Certs.EtcdClientCert); err != nil {
			return err
		}
		if err = apiServer.Flags().Set("etcd-keyfile", a.Certs.EtcdClientCertKey); err != nil {
			return err
		}
	}

	// Shutdown when context is closed
	go func() {
		<-ctx.Done()
		genericapiserver.RequestShutdown()
	}()

	return apiServer.ExecuteContext(ctx)
}

// GenerateKubeConfig will generate KubeConfig to allow access to cluster
func (a *APIServerConfig) GenerateKubeConfig(path string) error {

	caCertByte, err := ioutil.ReadFile(a.Certs.CACert)
	if err != nil {
		return fmt.Errorf("error read ca cert %v", err)
	}

	adminCertByte, err := ioutil.ReadFile(a.Certs.AdminCert)
	if err != nil {
		return fmt.Errorf("error read admin cert %v", err)
	}

	adminCertKeyByte, err := ioutil.ReadFile(a.Certs.AdminCertKey)
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

	context := kubeconfig.NewContext()
	context.AuthInfo = "default"
	context.Cluster = "default"

	kc.Clusters["default"] = cluster
	kc.AuthInfos["default"] = authInfo
	kc.Contexts["default"] = context
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

	a.Config = config

	return nil
}

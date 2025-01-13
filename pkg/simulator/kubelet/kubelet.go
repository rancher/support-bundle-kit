package kubelet

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/virtual-kubelet/node-cli/opts"
	"github.com/virtual-kubelet/virtual-kubelet/node/api"
	corev1 "k8s.io/api/core/v1"

	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/objects"
)

const (
	defaultStreamIdleTimeout = 10 * time.Second
)

type KubeletSimulator struct {
	ctx        context.Context
	certs      *certs.CertInfo
	bundlePath string
}

func NewKubeletSimulator(ctx context.Context, certs *certs.CertInfo, bundlePath string) (*KubeletSimulator, error) {
	// check bundle path exists

	if certs == nil {
		return nil, fmt.Errorf("no certificates provided for bootstrapping kubelet")
	}

	return &KubeletSimulator{
		ctx:        ctx,
		certs:      certs,
		bundlePath: bundlePath,
	}, nil
}
func (k *KubeletSimulator) RunFakeKubelet() error {
	routes := http.NewServeMux()
	podRoutes := api.PodHandlerConfig{
		RunInContainer:        k.runInContainer,
		GetContainerLogs:      k.getContainerLogs,
		GetPods:               k.getPods,
		StreamIdleTimeout:     defaultStreamIdleTimeout,
		StreamCreationTimeout: opts.DefaultStreamCreationTimeout,
	}

	api.AttachPodRoutes(podRoutes, routes, false)

	tlsConfig, err := loadTLSConfig(k.ctx, k.certs.KubeletCert, k.certs.KubeletCertKey, k.certs.CACert, true, false)
	if err != nil {
		return err

	}
	s := &http.Server{
		Handler:   routes,
		TLSConfig: tlsConfig,
	}

	l, err := tls.Listen("tcp", "127.0.0.1:10250", tlsConfig)
	if err != nil {
		return err
	}

	go func() {
		<-k.ctx.Done()
		if err := s.Shutdown(k.ctx); err != nil {
			log.Fatalf("error shutting down kubelet: %v", err)
		}
	}()

	return s.Serve(l)

}

// runInContainer does nothing
func (k *KubeletSimulator) runInContainer(ctx context.Context, namespace, name, container string, cmd []string, attach api.AttachIO) error {
	return nil
}

// getPods returns pod information
func (k *KubeletSimulator) getPods(ctx context.Context) ([]*corev1.Pod, error) {
	log.Printf("GetPods from path %s\n", k.bundlePath)
	return nil, nil
	//return util.GeneratePodList(k.bundlePath)
}

// getContainerLogs streams the logs from the bundle
func (k *KubeletSimulator) getContainerLogs(ctx context.Context, namespace, podName, containerName string, opts api.ContainerLogOpts) (io.ReadCloser, error) {
	log.Printf("Get logs for podName %s with opts %v\n", podName, opts)
	contents, err := readLogFiles(k.bundlePath, namespace, podName, containerName, opts.Previous)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(contents), nil
}

func readLogFiles(path, namespace, name, container string, loadPrevious bool) (io.Reader, error) {

	// node.zip files need to be handled using the zip file reader
	if namespace == objects.DefaultPodNamespace {
		return readZipFiles(path, name, container)
	}

	abs, err := filepath.Abs(filepath.Join(path, "logs", namespace, name, container+".log"))
	if loadPrevious {
		abs, err = filepath.Abs(filepath.Join(path, "logs", namespace, name, container+".log.1"))
	}
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(content), nil
}

func readZipFiles(path, name, container string) (io.Reader, error) {
	// generate zip file name
	abs, err := filepath.Abs(filepath.Join(path, "nodes", fmt.Sprintf("%s.zip", name)))

	if err != nil {
		return nil, err
	}

	_, err = os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("unable to find zip file corresponding to pod at path abs: %v", err)
	}

	r, err := zip.OpenReader(abs)

	if err != nil {
		return nil, fmt.Errorf("error reading node zip file %s: %v", name, err)
	}

	var content []byte
	var found bool
	for _, f := range r.File {
		if filepath.Base(f.Name) == fmt.Sprintf("%s.log", container) {
			content, err = objects.ReadContent(f)
			if err != nil {
				return nil, err
			}
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("could not find log file name %s.log", container)
	}
	return bytes.NewReader(content), r.Close()
}

func loadTLSConfig(ctx context.Context, certPath, keyPath, caPath string, allowUnauthenticatedClients, authWebhookEnabled bool) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("error loading tls certs %v", err)
	}

	var (
		AcceptedCiphers = []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,

			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		}
		caPool     *x509.CertPool
		clientAuth = tls.RequireAndVerifyClientCert
	)

	if allowUnauthenticatedClients {
		clientAuth = tls.NoClientCert
	}
	if authWebhookEnabled {
		clientAuth = tls.RequestClientCert
	}

	if caPath != "" {
		caPool = x509.NewCertPool()
		pem, err := os.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		if !caPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("error appending ca cert to certificate pool")
		}
	}

	return &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites:             AcceptedCiphers,
		ClientCAs:                caPool,
		ClientAuth:               clientAuth,
	}, nil
}

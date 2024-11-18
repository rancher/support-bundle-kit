package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	wranglerunstructured "github.com/rancher/wrangler/pkg/unstructured"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/rancher/support-bundle-kit/pkg/simulator/apiserver"
	"github.com/rancher/support-bundle-kit/pkg/simulator/certs"
	"github.com/rancher/support-bundle-kit/pkg/simulator/etcd"
	"github.com/rancher/support-bundle-kit/pkg/simulator/kubelet"
	"github.com/rancher/support-bundle-kit/pkg/simulator/objects"
)

var (
	simHome     string
	bundlePath  string
	resetHome   bool
	skipLoad    bool
	clientQPS   float32
	clientBurst int
)

var simulatorCmd = &cobra.Command{
	Use:   "simulator",
	Short: "Simulate a support bundle",
	Long: `Simulate a support bundle by loading into an empty apiserver
The simulator will 	run an embedded etcd, apiserver and a minimal virtual kubelet.
It will then load the support bundle into this setup, allowing users to browse and interact with
support bundle contents using native k8s tooling like kubectl`,
	Run: func(cmd *cobra.Command, args []string) {
		// clean home dir if needed
		var err error
		if resetHome {
			err = os.RemoveAll(simHome)
			if err != nil {
				logrus.Fatalf("Error during reset of sim-home: %v", err)
			}
		}

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		a := apiserver.NewAPIServerConfig(clientQPS, clientBurst)

		generatedCerts, err := certs.GenerateCerts([]string{"localhost"}, simHome)
		if err != nil {
			logrus.Fatalf("Error generating certificates %v", err)
		}
		a.Certs = generatedCerts

		etcdConfig, err := etcd.RunEmbeddedEtcd(ctx, filepath.Join(simHome), generatedCerts)
		if err != nil {
			logrus.Fatalf("Error setting up embedded etcdserver %v", err)
		}
		a.Etcd = etcdConfig

		err = a.GenerateKubeConfig(filepath.Join(simHome, "admin.kubeconfig"))
		if err != nil {
			logrus.Fatalf("Error generating kubeconfig %v", err)
		}

		eg, egctx := errgroup.WithContext(ctx)

		k, err := kubelet.NewKubeletSimulator(egctx, generatedCerts, bundlePath)
		if err != nil {
			logrus.Fatalf("Error initialisting kubelet simulator: %v", err)
		}

		serviceClusterIP, err := GetServiceClusterIP(bundlePath)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to get service cluster IP")
		}

		if serviceClusterIP == "" {
			serviceClusterIP = apiserver.DefaultServiceClusterIP
			logrus.Warnf("Cannot find service cluster IP, using default: %v", serviceClusterIP)
		}

		eg.Go(func() error {
			return a.RunAPIServer(egctx, serviceClusterIP)
		})

		eg.Go(func() error {
			return k.RunFakeKubelet()
		})

		o, err := objects.NewObjectManager(ctx, a.Config, bundlePath)
		if err != nil {
			logrus.Fatalf("Error creating object manager %v", err)
		}

		err = o.WaitForNamespaces(30 * time.Second)
		if err != nil {
			logrus.Fatal(err)
		}

		if !skipLoad {
			start := time.Now()
			err = o.CreateUnstructuredClusterObjects()

			if err != nil {
				logrus.Fatalf("Error loading cluster scoped objects %v", err)
			}

			err = o.CreateUnstructuredObjects()
			if err != nil {
				logrus.Fatalf("Error loading namespacedobjects %v", err)
			}

			err = o.CreateNodeZipObjects()
			if err != nil {
				logrus.Fatalf("Error loading node zip objects %v", err)
			}

			// ignore the error creation
			_ = o.CreatedFailedObjectsList()
			logrus.Info("All resources loaded successfully")

			logrus.Infof("Time to load all objects: %s seconds", time.Since(start))
		}

		err = eg.Wait()
		if err != nil {
			logrus.Fatalf("Error from apiserver or kublet subroutine: %v", err)
		}
	},
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		logrus.Fatalf("Error querying home directory %v", err)
	}

	dir := filepath.Join(home, ".sim")
	rootCmd.AddCommand(simulatorCmd)
	simulatorCmd.PersistentFlags().StringVar(&simHome, "sim-home", dir, "default home directory where sim stores its configuration. default is $HOME/.sim")
	simulatorCmd.PersistentFlags().StringVar(&bundlePath, "bundle-path", ".", "location to support bundle. default is .")
	simulatorCmd.PersistentFlags().BoolVar(&resetHome, "reset", false, "reset sim-home, will clear the contents and start a clean etcd + apiserver instance")
	simulatorCmd.PersistentFlags().BoolVar(&skipLoad, "skip-load", false, "skip load / re-load of bundle. this will ensure current etcd contents are only accessible")
	simulatorCmd.PersistentFlags().Float32Var(&clientQPS, "client-qps", apiserver.DefaultClientQPS, "QPS for the kubernete client to push objects to the api server")
	simulatorCmd.PersistentFlags().IntVar(&clientBurst, "client-burst", apiserver.DefaultClientBurst, "Burst for the kubernete client to push objects to the api server")
}

// GetServiceClusterIP will return the service cluster IP from the support bundle
func GetServiceClusterIP(bundlePath string) (string, error) {
	absPath, err := filepath.Abs(filepath.Join(bundlePath, "yamls/namespaced/default/v1/services.yaml"))
	if err != nil {
		return "", fmt.Errorf("failed to generate absolute path: %w", err)
	}

	objs, err := objects.GenerateObjects(absPath)
	if err != nil {
		return "", err
	}

	var kubeUnstructObj *unstructured.Unstructured
	for _, obj := range objs {
		unstructObj, err := wranglerunstructured.ToUnstructured(obj)
		if err != nil {
			return "", err
		}

		if unstructObj.GetName() == "kubernetes" {
			kubeUnstructObj = unstructObj
			break
		}

	}

	if kubeUnstructObj == nil {
		return "", fmt.Errorf("cannot find service kubernetes: %w", err)
	}

	clusterIP, ok, err := unstructured.NestedString(kubeUnstructObj.Object, "spec", "clusterIP")
	if err != nil {
		return "", fmt.Errorf("failed to fetch spec.clusterIP for service %v: %w", kubeUnstructObj.GetName(), err)
	}

	if !ok {
		return "", fmt.Errorf("could not find spec.clusterIP for service %s\n%v", kubeUnstructObj.GetName(), kubeUnstructObj)
	}

	return clusterIP, nil
}

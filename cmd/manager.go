package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rancher/support-bundle-kit/pkg/manager"
	"github.com/rancher/support-bundle-kit/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	sbm = &manager.SupportBundleManager{}
)

// managerCmd represents the manager command
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Harvester support bundle manager",
	Long: `Harvester support bundle manager

The manager collects following items:
- Cluster level bundle. Including resource manifests and pod logs.
- Any external bundles. e.g., Longhorn support bundle.

And it also waits for reports from support bundle agents. The reports contain:
- Logs of each Harvester node.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := sbm.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(managerCmd)
	managerCmd.PersistentFlags().StringVar(&sbm.NamespaceList, "namespaces", os.Getenv("NAMESPACES"), "List of namespaces delimited by ,")
	managerCmd.PersistentFlags().StringVar(&sbm.BundleName, "bundlename", os.Getenv("HARVESTER_SUPPORT_BUNDLE_NAME"), "The support bundle name")
	managerCmd.PersistentFlags().StringVar(&sbm.OutputDir, "outdir", os.Getenv("HARVESTER_SUPPORT_BUNDLE_OUTPUT_DIR"), "The directory to store the bundle")

	timeout := utils.EnvGetDuration("HARVESTER_SUPPORT_BUNDLE_WAIT_TIMEOUT", 5*time.Minute)
	managerCmd.PersistentFlags().DurationVar(&sbm.WaitTimeout, "wait", timeout, "The timeout to wait for node bundles")

	managerCmd.PersistentFlags().StringVar(&sbm.LonghornAPI, "longhorn-api", "http://longhorn-backend.longhorn-system:9500", "The Longhorn API URL")
	managerCmd.PersistentFlags().StringVar(&sbm.ManagerPodIP, "manager-pod-ip", os.Getenv("HARVESTER_SUPPORT_BUNDLE_MANAGER_POD_IP"), "The Harvester support bundle manager's IP (pod runs this app)")
	managerCmd.PersistentFlags().StringVar(&sbm.ImageName, "image-name", os.Getenv("HARVESTER_SUPPORT_BUNDLE_IMAGE"), "The Harvester support bundle image")
	managerCmd.PersistentFlags().StringVar(&sbm.ImagePullPolicy, "image-pull-policy", os.Getenv("HARVESTER_SUPPORT_BUNDLE_IMAGE_PULL_POLICY"), "Pull policy of the Harvester support bundle image")

	managerCmd.PersistentFlags().BoolVar(&sbm.Standalone, "standalone", false, "Run the manager in standalone mode. Harvester api server is not required.")
}

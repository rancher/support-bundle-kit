package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	AppVersion = "dev"
	GitCommit  = "commit"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get application version",
	Long:  "Get application version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s (%s)\n", AppVersion, GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

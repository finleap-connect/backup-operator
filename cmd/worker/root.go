package main

import (
	"flag"
	"os"

	"github.com/kubism-io/backup-operator/pkg/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "worker action [flags]",
	Short:        "Worker provides several backup implementations, which are run in CronJobs.",
	SilenceUsage: true,
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.AddGoFlagSet(flag.CommandLine)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log := logger.WithName("root-cmd")
		log.Error(err, "command failed")
		os.Exit(1)
	}
}

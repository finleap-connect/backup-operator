package main

import (
	"github.com/spf13/cobra"
)

var mongodbCmd = &cobra.Command{
	Use:   "mongodb [flags] config",
	Short: "Backups mongodb using specified config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mongodbCmd)
}

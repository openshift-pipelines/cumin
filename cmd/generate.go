package cmd

import (
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate manifests for the dashboard",
}

func init() {
	dashboardCmd.AddCommand(generateCmd)
}

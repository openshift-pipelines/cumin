package cmd

import (
	"github.com/spf13/cobra"
)

// dashboardCmd represents the dashboard command
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "dashboard foo",
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

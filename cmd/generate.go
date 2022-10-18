package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var (
	jenkinsUsername string
	jenkinsPassword string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate manifests for the dashboard",
}

func init() {
	dashboardCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringVarP(&jenkinsUsername, "username", "u", os.Getenv("JENKINS_USERNAME"), "jenkins username")
	generateCmd.PersistentFlags().StringVarP(&jenkinsPassword, "password", "p", os.Getenv("JENKINS_PASSWORD"), "jenkins password")
}

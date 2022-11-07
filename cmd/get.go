package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get verb for jira",
}

func init() {
	jiraCmd.AddCommand(getCmd)
}

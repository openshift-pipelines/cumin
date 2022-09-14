package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jiraBaseURL string
	jiraBoardID int
	jiraProject string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cumin",
	Short: "jira foo",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&jiraBaseURL, "base", "", "jira base url")
	if err := rootCmd.MarkPersistentFlagRequired("base"); err != nil {
		return
	}

	rootCmd.PersistentFlags().IntVar(&jiraBoardID, "board-id", 0, "jira board id")
	if err := rootCmd.MarkPersistentFlagRequired("board-id"); err != nil {
		return
	}

	rootCmd.PersistentFlags().StringVarP(&jiraProject, "project", "p", "", "jira project to clone the issue into")
	if err := rootCmd.MarkFlagRequired("project"); err != nil {
		return
	}
}

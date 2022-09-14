package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	rootCmd.PersistentFlags().StringVar(&jiraBaseURL, "base", os.Getenv("JIRA_BASE_URL"), "jira base url (env: JIRA_BASE_URL)")
	if err := rootCmd.MarkPersistentFlagRequired("base"); err != nil {
		log.Fatal(err)
	}

	defaultJiraBoardID := 0
	if os.Getenv("JIRA_BOARD_ID") != "" {
		var err error
		if defaultJiraBoardID, err = strconv.Atoi(os.Getenv("JIRA_BOARD_ID")); err != nil {
			log.Fatal(err)
		}
	}
	rootCmd.PersistentFlags().IntVar(&jiraBoardID, "board-id", defaultJiraBoardID, "jira board id (env: JIRA_BOARD_ID)")
	if err := rootCmd.MarkPersistentFlagRequired("board-id"); err != nil {
		return
	}

	rootCmd.PersistentFlags().StringVarP(&jiraProject, "project", "p", os.Getenv("JIRA_PROJECT"), "jira project to clone the issue into (env: JIRA_PROJECT)")
	if err := rootCmd.MarkFlagRequired("project"); err != nil {
		log.Fatal(err)
	}
}

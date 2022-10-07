package cmd

import (
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

// jiraCmd represents the jira command
var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "jira foo",
}

func init() {
	rootCmd.AddCommand(jiraCmd)

	jiraCmd.PersistentFlags().StringVar(&jiraBaseURL, "base", os.Getenv("JIRA_BASE_URL"), "jira base url (env: JIRA_BASE_URL)")
	if err := jiraCmd.MarkPersistentFlagRequired("base"); err != nil {
		log.Fatal(err)
	}

	defaultJiraBoardID := 0
	if os.Getenv("JIRA_BOARD_ID") != "" {
		var err error
		if defaultJiraBoardID, err = strconv.Atoi(os.Getenv("JIRA_BOARD_ID")); err != nil {
			log.Fatal(err)
		}
	}
	jiraCmd.PersistentFlags().IntVar(&jiraBoardID, "board-id", defaultJiraBoardID, "jira board id (env: JIRA_BOARD_ID)")
	if err := jiraCmd.MarkPersistentFlagRequired("board-id"); err != nil {
		return
	}

	jiraCmd.PersistentFlags().StringVarP(&jiraProject, "project", "p", os.Getenv("JIRA_PROJECT"), "jira project to clone the issue into (env: JIRA_PROJECT)")
	if err := jiraCmd.MarkPersistentFlagRequired("project"); err != nil {
		log.Fatal(err)
	}
}

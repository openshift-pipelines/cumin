package cmd

import (
	"context"
	"fmt"
	"github.com/concaf/cumin/pkg/dashboard"
	"github.com/concaf/cumin/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
)

// generateIssuesCmd represents the issues command
var generateIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "generate issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		jiraConfig := shared.JiraConfig{
			BaseURL: jiraBaseURL,
			BoardID: jiraBoardID,
			Project: jiraProject,
		}

		ctx := context.Background()
		jiraClient, err := shared.NewJiraClient(jiraBaseURL)
		if err != nil {
			return err
		}

		issueList, err := dashboard.GenerateIssueList(ctx, jql, jiraClient, &jiraConfig)
		if err != nil {
			return err
		}

		marshalledIssueList, err := yaml.Marshal(issueList)
		if err != nil {
			return err
		}

		fmt.Println(string(marshalledIssueList))
		return nil
	},
}

func init() {
	generateCmd.AddCommand(generateIssuesCmd)

	generateIssuesCmd.PersistentFlags().StringVar(&jiraBaseURL, "base", os.Getenv("JIRA_BASE_URL"), "jira base url (env: JIRA_BASE_URL)")
	if err := generateIssuesCmd.MarkPersistentFlagRequired("base"); err != nil {
		log.Fatal(err)
	}

	defaultJiraBoardID := 0
	if os.Getenv("JIRA_BOARD_ID") != "" {
		var err error
		if defaultJiraBoardID, err = strconv.Atoi(os.Getenv("JIRA_BOARD_ID")); err != nil {
			log.Fatal(err)
		}
	}
	generateIssuesCmd.PersistentFlags().IntVar(&jiraBoardID, "board-id", defaultJiraBoardID, "jira board id (env: JIRA_BOARD_ID)")
	if err := generateIssuesCmd.MarkPersistentFlagRequired("board-id"); err != nil {
		return
	}

	generateIssuesCmd.PersistentFlags().StringVar(&jiraProject, "project", os.Getenv("JIRA_PROJECT"), "jira project to clone the issue into (env: JIRA_PROJECT)")
	if err := generateIssuesCmd.MarkPersistentFlagRequired("project"); err != nil {
		log.Fatal(err)
	}
	generateIssuesCmd.Flags().StringVar(&jql, "jql", "", "jql to list issues")
	generateIssuesCmd.MarkFlagRequired("jql")
}

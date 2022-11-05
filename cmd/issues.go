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

// issuesCmd represents the issues command
var issuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	generateCmd.AddCommand(issuesCmd)

	issuesCmd.PersistentFlags().StringVar(&jiraBaseURL, "base", os.Getenv("JIRA_BASE_URL"), "jira base url (env: JIRA_BASE_URL)")
	if err := issuesCmd.MarkPersistentFlagRequired("base"); err != nil {
		log.Fatal(err)
	}

	defaultJiraBoardID := 0
	if os.Getenv("JIRA_BOARD_ID") != "" {
		var err error
		if defaultJiraBoardID, err = strconv.Atoi(os.Getenv("JIRA_BOARD_ID")); err != nil {
			log.Fatal(err)
		}
	}
	issuesCmd.PersistentFlags().IntVar(&jiraBoardID, "board-id", defaultJiraBoardID, "jira board id (env: JIRA_BOARD_ID)")
	if err := issuesCmd.MarkPersistentFlagRequired("board-id"); err != nil {
		return
	}

	issuesCmd.PersistentFlags().StringVar(&jiraProject, "project", os.Getenv("JIRA_PROJECT"), "jira project to clone the issue into (env: JIRA_PROJECT)")
	if err := issuesCmd.MarkPersistentFlagRequired("project"); err != nil {
		log.Fatal(err)
	}
	issuesCmd.Flags().StringVar(&jql, "jql", "", "jql to list issues")
	issuesCmd.MarkFlagRequired("jql")
}

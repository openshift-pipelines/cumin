package cmd

import (
	"context"
	"fmt"
	"github.com/concaf/cumin/pkg/jira"
	"github.com/concaf/cumin/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
)

var (
	jql string
)

// getIssuesCmd represents the list command
var getIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "get issues per jql",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		jiraConfig := shared.JiraConfig{
			BaseURL: jiraBaseURL,
			BoardID: jiraBoardID,
			Project: jiraProject,
		}

		jiraClient, err := shared.NewJiraClient(jiraBaseURL)
		if err != nil {
			return err
		}

		issues, err := jira.ListIssues(ctx, jiraClient, &jiraConfig, jql)
		if err != nil {
			return err
		}

		issuesYaml, err := yaml.Marshal(issues)
		if err != nil {
			return err
		}

		fmt.Println(string(issuesYaml))
		return nil
	},
}

func init() {
	getCmd.AddCommand(getIssuesCmd)

	getIssuesCmd.Flags().StringVar(&jql, "jql", "", "query in jql")
	if err := getIssuesCmd.MarkFlagRequired("jql"); err != nil {
		log.Println(err)
		return
	}
}

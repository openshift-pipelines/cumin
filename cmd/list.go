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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
	jiraCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&jql, "jql", "", "query in jql")
	if err := listCmd.MarkFlagRequired("jql"); err != nil {
		log.Println(err)
		return
	}
}

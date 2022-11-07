package cmd

import (
	"context"
	"fmt"
	"github.com/concaf/cumin/pkg/jira"
	"github.com/concaf/cumin/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	activeSprint bool
)

// sprintCmd represents the sprint command
var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "get sprint info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if activeSprint {
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

			sprint, err := jira.GetActiveSprint(ctx, jiraClient, &jiraConfig)
			if err != nil {
				return err
			}

			sprintYaml, err := yaml.Marshal(sprint)
			if err != nil {
				return err
			}

			fmt.Println(string(sprintYaml))
			return nil
		}
		return fmt.Errorf("i don't know what to do for a non-active sprint yet")
	},
}

func init() {
	getCmd.AddCommand(sprintCmd)

	sprintCmd.Flags().BoolVar(&activeSprint, "active", true, "get active sprint")
}

package cmd

import (
	"context"
	"errors"
	"github.com/concaf/cumin/pkg/clone"
	"github.com/concaf/cumin/pkg/shared"
	"github.com/spf13/cobra"
	"log"
)

var (
	labels             []string
	issueType          string
	fixVersions        []string
	priority           string
	assignee           string
	addToCurrentSprint bool
	storyPoints        int
	title              string
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "clone an issue from github to jira",
	Long: `clone an issue from github to jira
usage:

cumin clone https://github.com/concaf/cumin/issues/2455 \
--project SRVKP \
--label "imported-from-github" \
--label "groomable" \
--type story \
--fix-version "Pipelines 1.10" \
--priority critical \
--assignee concaf \
--add-to-current-sprint \
--story-points 5 \
--title "this cool upstream issue"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("no github issues specified to clone")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		jiraConfig := shared.JiraConfig{
			BaseURL: jiraBaseURL,
			BoardID: jiraBoardID,
			Project: jiraProject,
		}

		jiraIssueSchema := shared.JiraIssueSchema{
			Project:            jiraProject,
			Labels:             labels,
			Type:               issueType,
			FixVersions:        fixVersions,
			Priority:           priority,
			Assignee:           assignee,
			AddToCurrentSprint: addToCurrentSprint,
			StoryPoints:        storyPoints,
			Title:              title,
		}

		for _, issue := range args {
			ghIssueSchema, err := shared.GitHubIssueSchemaFromURL(issue)
			if err != nil {
				return err
			}

			createdIssue, err := clone.Clone(ctx, &jiraConfig, ghIssueSchema, &jiraIssueSchema)
			if err != nil {
				return err
			}

			log.Printf("created issue: %v", createdIssue)
		}
		return nil
	},
}

func init() {
	jiraCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringArrayVarP(&labels, "labels", "l", nil, "labels to add to the jira issue")

	cloneCmd.Flags().StringVarP(&issueType, "type", "t", "story", "type of the jira issue, e.g. story, bug, etc")
	if err := cloneCmd.MarkFlagRequired("type"); err != nil {
		log.Println(err)
		return
	}

	cloneCmd.Flags().StringArrayVar(&fixVersions, "fix-versions", nil, "fixVersion(s) to set on the jira issue")

	cloneCmd.Flags().StringVar(&priority, "priority", "", "priority to set on the jira issue, e.g. major, critical, blocker, etc")

	cloneCmd.Flags().StringVarP(&assignee, "assignee", "a", "", "assignee to set on the jira issue")

	cloneCmd.Flags().BoolVar(&addToCurrentSprint, "add-to-current-sprint", false, "add the jira issue to current sprint?")

	cloneCmd.Flags().IntVar(&storyPoints, "story-points", 0, "story points to add to the jira issue")

	cloneCmd.Flags().StringVar(&title, "title", "", "override the title in the jira issue")
}

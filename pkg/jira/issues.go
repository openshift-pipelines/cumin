package jira

import (
	"context"
	"github.com/andygrunwald/go-jira"
	"github.com/concaf/cumin/pkg/shared"
)

func ListIssues(ctx context.Context, jiraClient *jira.Client, jiraConfig *shared.JiraConfig, jql string) ([]shared.JiraIssueSchema, error) {
	issues, _, err := jiraClient.Issue.Search(jql, &jira.SearchOptions{MaxResults: 1000})
	if err != nil {
		return nil, err
	}
	var jiraIssues []shared.JiraIssueSchema
	for _, issue := range issues {
		jiraIssue := shared.JiraIssueSchema{
			Key:         issue.Key,
			Type:        issue.Fields.Type.Name,
			Project:     issue.Fields.Project.Key,
			Priority:    issue.Fields.Priority.Name,
			Status:      issue.Fields.Status.Name,
			Labels:      issue.Fields.Labels,
			Description: issue.Fields.Description,
			Title:       issue.Fields.Summary,
		}
		if issue.Fields.Assignee != nil {
			jiraIssue.Assignee = shared.JiraUser{
				Name:        issue.Fields.Assignee.Name,
				Email:       issue.Fields.Assignee.EmailAddress,
				DisplayName: issue.Fields.Assignee.DisplayName,
				Key:         issue.Fields.Assignee.Key,
			}
		}
		if issue.Fields.Unknowns[shared.StoryPointsCustomField] != nil {
			jiraIssue.StoryPoints = issue.Fields.Unknowns[shared.StoryPointsCustomField].(float64)
		}
		var fixVersions []string
		for _, fixVersion := range issue.Fields.FixVersions {
			fixVersions = append(fixVersions, fixVersion.Name)
		}
		jiraIssue.FixVersions = fixVersions

		jiraIssues = append(jiraIssues, jiraIssue)
	}
	return jiraIssues, nil
}

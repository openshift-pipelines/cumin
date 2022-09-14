package shared

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/trivago/tgo/tcontainer"
	"strings"
)

type JiraConfig struct {
	BaseURL string
	BoardID int
	Project string
}

type JiraIssueSchema struct {
	Project            string
	Description        string
	Labels             []string
	Type               string
	FixVersions        []string
	Priority           string
	Assignee           string
	AddToCurrentSprint bool
	StoryPoints        int
	Title              string
}

func SchemaToJiraIssue(schema *JiraIssueSchema, currentSprintID int) (*jira.Issue, error) {
	unknowns := tcontainer.NewMarshalMap()

	if schema.Title == "" {
		return nil, fmt.Errorf("no title provided for the issue")
	}
	issue := jira.Issue{
		Fields: &jira.IssueFields{
			Summary: schema.Title,
		},
	}

	if schema.Project != "" {
		issue.Fields.Project = jira.Project{
			Key: schema.Project,
		}
	}

	issue.Fields.Labels = schema.Labels

	if schema.Type != "" {
		issue.Fields.Type = jira.IssueType{
			Name: schema.Type,
		}
	}

	var jiraFixVersions []*jira.FixVersion
	for _, fv := range schema.FixVersions {
		jiraFixVersions = append(jiraFixVersions, &jira.FixVersion{
			Name: fv,
		})
	}
	issue.Fields.FixVersions = jiraFixVersions

	if schema.Priority != "" {
		issue.Fields.Priority = &jira.Priority{
			Name: schema.Priority,
		}
	}

	if schema.Assignee != "" {
		issue.Fields.Assignee = &jira.User{
			Name: schema.Assignee,
		}
	}

	if schema.AddToCurrentSprint {
		if currentSprintID == 0 {
			return nil, fmt.Errorf("no sprint id specified to add to current sprint")
		}
		unknowns["customfield_12310940"] = currentSprintID
	}

	if schema.StoryPoints != 0 {
		unknowns["customfield_12310243"] = schema.StoryPoints
	}

	if schema.Description != "" {
		issue.Fields.Description = schema.Description
	}

	if len(unknowns) > 0 {
		issue.Fields.Unknowns = unknowns
	}

	return &issue, nil
}

func NewJiraClient(token string, baseURL string) (*jira.Client, error) {
	jiraAuthTransport := jira.BearerAuthTransport{
		Token: strings.TrimSpace(token),
	}
	jiraClient, err := jira.NewClient(jiraAuthTransport.Client(), baseURL)
	if err != nil {
		return nil, err
	}
	return jiraClient, nil
}

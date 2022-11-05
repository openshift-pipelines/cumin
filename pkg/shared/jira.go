package shared

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/trivago/tgo/tcontainer"
	"log"
	"os"
	"strings"
)

type JiraConfig struct {
	BaseURL string
	BoardID int
	Project string
}

type JiraIssueSchema struct {
	Key                string // SRVKP-1337
	Type               string // bug, etc
	Project            string // SRVKP
	Priority           string
	Status             string
	FixVersions        []string
	Labels             []string
	Description        string
	Assignee           JiraUser
	StoryPoints        float64
	Title              string
	AddToCurrentSprint bool
}

type JiraUser struct {
	Name        string
	Email       string
	DisplayName string
	Key         string
}

type JiraVersion struct {
	Name     string
	Released bool
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

	if schema.Assignee.Key != "" {
		issue.Fields.Assignee = &jira.User{
			Name: schema.Assignee.Key,
		}
	}

	if schema.AddToCurrentSprint {
		if currentSprintID == 0 {
			return nil, fmt.Errorf("no sprint id specified to add to current sprint")
		}
		unknowns["customfield_12310940"] = currentSprintID
	}

	if schema.StoryPoints != 0 {
		unknowns[StoryPointsCustomField] = schema.StoryPoints
	}

	if schema.Description != "" {
		issue.Fields.Description = schema.Description
	}

	if len(unknowns) > 0 {
		issue.Fields.Unknowns = unknowns
	}

	return &issue, nil
}

func NewJiraClient(baseURL string) (*jira.Client, error) {
	token, ok := os.LookupEnv(JiraTokenEnvVar)
	if !ok {
		return nil, fmt.Errorf("jira token not set in environment variable: %v", JiraTokenEnvVar)
	}
	token = strings.TrimSpace(token)

	jiraAuthTransport := jira.BearerAuthTransport{
		Token: strings.TrimSpace(token),
	}
	jiraClient, err := jira.NewClient(jiraAuthTransport.Client(), baseURL)
	if err != nil {
		return nil, err
	}
	return jiraClient, nil
}

func GetVersions(jiraConfig *JiraConfig) ([]JiraVersion, error) {
	jiraClient, err := NewJiraClient(jiraConfig.BaseURL)
	if err != nil {
		return nil, err
	}
	log.Println("jira client generation successful")

	project, _, err := jiraClient.Project.Get(jiraConfig.Project)
	if err != nil {
		return nil, err
	}

	var versions []JiraVersion
	for _, version := range project.Versions {
		versions = append(versions, JiraVersion{
			Name:     version.Name,
			Released: *version.Released,
		})
	}
	return versions, nil
}

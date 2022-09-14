package clone

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/concaf/cumin/pkg/shared"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

func Clone(ctx context.Context, jiraConfig *shared.JiraConfig, ghIssueSchema *shared.GitHubIssueSchema, jiraIssueSchema *shared.JiraIssueSchema) (string, error) {
	// get issue from github
	ghToken, ok := os.LookupEnv(shared.GitHubTokenEnvVar)
	if ok {
		log.Printf("found github token set in environment variable %v, using...\n", shared.GitHubTokenEnvVar)
	} else {
		log.Printf("github token is not set in the environment variable %v, skipping...\n", shared.GitHubTokenEnvVar)
	}

	ghClient := shared.NewGitHubClient(ctx, ghToken)
	log.Println("github client generation successful")

	jiraToken, ok := os.LookupEnv(shared.JiraTokenEnvVar)
	if !ok {
		return "", fmt.Errorf("jira token not set in environment variable: %v", shared.JiraTokenEnvVar)
	}
	jiraToken = strings.TrimSpace(jiraToken)

	jiraClient, err := shared.NewJiraClient(jiraToken, jiraConfig.BaseURL)
	if err != nil {
		return "", err
	}
	log.Println("jira client generation successful")

	log.Printf("fetching github issue...")
	ghIssue, _, err := ghClient.Issues.Get(ctx, ghIssueSchema.Owner, ghIssueSchema.Repo, ghIssueSchema.Number)
	if err != nil {
		return "", err
	}
	log.Printf("github issue fetch successful")

	if jiraIssueSchema.Title == "" {
		jiraIssueSchema.Title = *ghIssue.Title
	} else {
		log.Printf("title supplied via cli, github issue title with be overwritten...")
	}
	log.Printf("jira issue title: %v", jiraIssueSchema.Title)

	// set jira issue description as github issue body
	jiraIssueSchema.Description = fmt.Sprintf("%v\n"+
		"{quote}\n"+
		"This issue was cloned from the GitHub issue %v on %v\n"+
		"Powered by [cumin|https://github.com/concaf/cumin]\n"+
		"{quote}", *ghIssue.Body, ghIssueSchema.URL, time.Now().Format("02 January 2006 3:4:5 PM MST (-0700)"))

	var currentSprintID int
	if jiraIssueSchema.AddToCurrentSprint {
		log.Println("the issue will be added to the current sprint")
		sprintsList, _, err := jiraClient.Board.GetAllSprintsWithOptions(jiraConfig.BoardID, &jira.GetAllSprintsOptions{
			State: "active",
		})
		if err != nil {
			return "", err
		}
		// we don't want more than one active sprints
		if len(sprintsList.Values) != 1 {
			return "", fmt.Errorf("expected exactly one active sprint, got: %v", len(sprintsList.Values))
		}

		log.Printf("found one active sprint: %v", sprintsList.Values[0].Name)
		currentSprintID = sprintsList.Values[0].ID
	}

	jiraIssue, err := shared.SchemaToJiraIssue(jiraIssueSchema, currentSprintID)
	if err != nil {
		return "", err
	}

	log.Println("creating jira issue now...")
	createdIssue, resp, err := jiraClient.Issue.Create(jiraIssue)
	if err != nil {
		if resp != nil && resp.Body != nil {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("error reading the response body: %v", err)
			} else {
				log.Printf("response body: %v", string(body))
			}
		}
		return "", err
	}

	log.Printf("created issue id: %v", createdIssue.Key)

	issueURL, err := url.JoinPath(jiraConfig.BaseURL, "browse", createdIssue.Key)
	if err != nil {
		return "", err
	}
	return issueURL, nil
}

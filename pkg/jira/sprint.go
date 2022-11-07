package jira

import (
	"context"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/concaf/cumin/pkg/shared"
)

func GetActiveSprint(ctx context.Context, jiraClient *jira.Client, jiraConfig *shared.JiraConfig) (*jira.Sprint, error) {
	activeSprints, _, err := jiraClient.Board.GetAllSprintsWithOptions(jiraConfig.BoardID, &jira.GetAllSprintsOptions{
		State: "active",
	})
	if err != nil {
		return nil, err
	}
	if len(activeSprints.Values) != 1 {
		// we don't know how to deal with this yet
		return nil, fmt.Errorf("expected exactly one active sprint, got %v", len(activeSprints.Values))
	}
	return &activeSprints.Values[0], nil
}

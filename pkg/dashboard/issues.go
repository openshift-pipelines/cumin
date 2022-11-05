package dashboard

import (
	"context"
	jiraLib "github.com/andygrunwald/go-jira"
	"github.com/concaf/cumin/pkg/jira"
	"github.com/concaf/cumin/pkg/shared"
	"log"
	"strings"
)

type DashboardIssueList struct {
	StoryPoints struct {
		Total          float64
		New            float64
		DevComplete    float64
		ReleasePending float64
		ToDo           float64
		InProgress     float64
		CodeReview     float64
		OnQA           float64
		Verified       float64
		Closed         float64
	}
	Issues        []shared.JiraIssueSchema
	PatchReleases []*PatchRelease
}

type PatchRelease struct {
	Name        string
	StoryPoints float64
	Released    bool
	Issues      []shared.JiraIssueSchema
}

func GenerateIssueList(ctx context.Context, jql string, jiraClient *jiraLib.Client, jiraConfig *shared.JiraConfig) (*DashboardIssueList, error) {
	// get issues for version
	issues, err := jira.ListIssues(ctx, jiraClient, jiraConfig, jql)
	if err != nil {
		return nil, err
	}

	var dashboardIssueList DashboardIssueList
	dashboardIssueList.Issues = issues
	for _, issue := range issues {
		log.Printf("generating for issue - %v, with status %v", issue.Key, issue.Status)

		// add the story points
		dashboardIssueList.StoryPoints.Total = dashboardIssueList.StoryPoints.Total + issue.StoryPoints
		normalizedStatus := strings.ReplaceAll(strings.ToLower(issue.Status), " ", "")
		if normalizedStatus == "todo" {
			dashboardIssueList.StoryPoints.ToDo = dashboardIssueList.StoryPoints.ToDo + issue.StoryPoints
		} else if normalizedStatus == "inprogress" {
			dashboardIssueList.StoryPoints.InProgress = dashboardIssueList.StoryPoints.InProgress + issue.StoryPoints
		} else if normalizedStatus == "codereview" {
			dashboardIssueList.StoryPoints.CodeReview = dashboardIssueList.StoryPoints.CodeReview + issue.StoryPoints
		} else if normalizedStatus == "onqa" {
			dashboardIssueList.StoryPoints.OnQA = dashboardIssueList.StoryPoints.OnQA + issue.StoryPoints
		} else if normalizedStatus == "verified" {
			dashboardIssueList.StoryPoints.Verified = dashboardIssueList.StoryPoints.Verified + issue.StoryPoints
		} else if normalizedStatus == "closed" {
			dashboardIssueList.StoryPoints.Closed = dashboardIssueList.StoryPoints.Closed + issue.StoryPoints
		} else if normalizedStatus == "devcomplete" {
			dashboardIssueList.StoryPoints.DevComplete = dashboardIssueList.StoryPoints.DevComplete + issue.StoryPoints
		} else if normalizedStatus == "releasepending" {
			dashboardIssueList.StoryPoints.ReleasePending = dashboardIssueList.StoryPoints.ReleasePending + issue.StoryPoints
		} else if normalizedStatus == "new" {
			dashboardIssueList.StoryPoints.New = dashboardIssueList.StoryPoints.New + issue.StoryPoints
		} else {
			log.Printf("invalid (?) status (%v) on issue %v, skipping...", issue.Status, issue.Key)
		}

		for _, fv := range issue.FixVersions {
			// fv == "Pipelines 1.8.1"
			splitFv := strings.Split(fv, " ")
			// version == "1.8.1"
			version := splitFv[len(splitFv)-1]

			var issuePatchRelease *PatchRelease
			// if patch release doesn't exist, create it
			patchReleaseExists := false
			for _, pr := range dashboardIssueList.PatchReleases {
				if version == pr.Name {
					patchReleaseExists = true
					issuePatchRelease = pr
					break
				}
			}
			if !patchReleaseExists {
				jiraVersion, err := shared.GetVersion(jiraClient, jiraConfig, fv)
				if err != nil {
					return nil, err
				}
				issuePatchRelease = &PatchRelease{
					Name:     version,
					Released: jiraVersion.Released,
				}
				dashboardIssueList.PatchReleases = append(dashboardIssueList.PatchReleases, issuePatchRelease)
			}

			issuePatchRelease.Issues = append(issuePatchRelease.Issues, issue)
			issuePatchRelease.StoryPoints = issuePatchRelease.StoryPoints + issue.StoryPoints
		}
	}
	return &dashboardIssueList, nil
}

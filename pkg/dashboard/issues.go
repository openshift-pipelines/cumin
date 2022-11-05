package dashboard

import (
	"context"
	"fmt"
	jiraLib "github.com/andygrunwald/go-jira"
	"github.com/concaf/cumin/pkg/jira"
	"github.com/concaf/cumin/pkg/shared"
	"github.com/coreos/go-semver/semver"
	"log"
	"strings"
)

type DashboardIssueList struct {
	MinorReleases []*MinorRelease
}

type MinorRelease struct {
	Name          string
	PatchReleases []PatchRelease
}

type PatchRelease struct {
	Name        string
	StoryPoints float64
	Released    bool
	Issues      []shared.JiraIssueSchema
}

func GenerateIssueList(ctx context.Context, jql string, jiraClient *jiraLib.Client, jiraConfig *shared.JiraConfig) (*DashboardIssueList, error) {
	versions, err := shared.GetVersions(jiraConfig)
	if err != nil {
		return nil, err
	}

	var minorReleases []*MinorRelease
	for _, version := range versions {
		log.Printf("generating for version %v ...", version.Name)
		var patchRelease PatchRelease

		// version == "Pipelines 1.8.1"
		versionSplit := strings.Split(version.Name, " ")
		patchVersion := versionSplit[len(versionSplit)-1]
		if strings.Count(patchVersion, ".") != 2 {
			log.Printf("patch version (%v) doesn't look right, should be x.y.z; ignoring...", patchVersion)
			continue
		}

		versionObj, err := semver.NewVersion(patchVersion)
		if err != nil {
			log.Printf("something went wrong with parsing version %v, ignoring...", patchVersion)
			log.Println(err)
			continue
		}

		// let's build the patch release
		patchRelease.Name = patchVersion
		patchRelease.Released = version.Released

		// get issues for version
		issues, err := jira.ListIssues(ctx, jiraClient, jiraConfig, jql)
		if err != nil {
			return nil, err
		}
		patchRelease.Issues = issues

		// work out story points now
		var storyPoints float64
		for _, issue := range issues {
			storyPoints = storyPoints + issue.StoryPoints
		}
		patchRelease.StoryPoints = storyPoints

		// add to the relevant minor version, create if does not exist
		minorVersion := fmt.Sprintf("%v.%v", versionObj.Major, versionObj.Minor)

		foundMinorRelease := false
		for _, existingMinorRelease := range minorReleases {
			if existingMinorRelease.Name == minorVersion {
				foundMinorRelease = true
				existingMinorRelease.PatchReleases = append(existingMinorRelease.PatchReleases, patchRelease)
				break
			}
		}

		// create a new minor release if not found
		if !foundMinorRelease {
			minorReleases = append(minorReleases, &MinorRelease{
				Name: minorVersion,
				PatchReleases: []PatchRelease{
					patchRelease,
				},
			})
		}
	}

	return &DashboardIssueList{
		MinorReleases: minorReleases,
	}, nil
}

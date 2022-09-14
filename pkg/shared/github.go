package shared

import (
	"context"
	"fmt"
	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type GitHubIssueSchema struct {
	URL    string
	Owner  string
	Repo   string
	Number int
}

func GitHubIssueSchemaFromURL(urlString string) (*GitHubIssueSchema, error) {
	issueURL, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	pathSplit := strings.Split(issueURL.Path, "/")[1:]

	if len(pathSplit) != 4 {
		return nil, fmt.Errorf("url path expected to be of the format <owner>/<repo>/issues/<number>, got: %v", issueURL.Path)
	}
	if pathSplit[2] != "issues" {
		return nil, fmt.Errorf("URL is not a github issue, expected format <owner>/<repo>/issues/<number>, got: %v", issueURL.Path)
	}
	issueNumber, err := strconv.Atoi(pathSplit[3])
	if err != nil {
		return nil, err
	}

	return &GitHubIssueSchema{
		URL:    urlString,
		Owner:  pathSplit[0],
		Repo:   pathSplit[1],
		Number: issueNumber,
	}, nil
}

func NewGitHubClient(ctx context.Context, token string) *github.Client {
	var httpClient *http.Client = nil
	// if token exists, use it
	if token != "" {
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token,
		})
		httpClient = oauth2.NewClient(ctx, tokenSource)
	}
	return github.NewClient(httpClient)
}

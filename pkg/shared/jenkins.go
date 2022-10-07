package shared

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type JenkinsBuild struct {
	Actions []struct {
		Causes []struct {
			ShortDescription string `json:"shortDescription"`
		} `json:"causes"`
	} `json:"actions"`
	Building        bool   `json:"building"`
	Duration        int    `json:"duration"`
	FullDisplayName string `json:"fullDisplayName"`
	ID              string `json:"id"`
	Number          int    `json:"number"`
	Result          string `json:"result"`
	Timestamp       int64  `json:"timestamp"`
	URL             string `json:"url"`
	PreviousBuild   struct {
		Number int    `json:"number"`
		URL    string `json:"url"`
	} `json:"previousBuild"`
}

func GetBuildJSONFromURL(buildURL, username, password string) (*JenkinsBuild, error) {
	buildURL, err := url.JoinPath(buildURL, "/api/json")
	if err != nil {
		return nil, err
	}
	log.Printf("fetching json from %v", buildURL)

	client := &http.Client{
		Timeout: time.Second * 20,
	}
	req, err := http.NewRequest("GET", buildURL, nil)
	if err != nil {
		return nil, err
	}

	// set up basic auth
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var build JenkinsBuild
	if err := json.Unmarshal(body, &build); err != nil {
		return nil, err
	}

	return &build, nil
}

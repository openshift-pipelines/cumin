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
		Class      string `json:"_class,omitempty"`
		Parameters []struct {
			Class string      `json:"_class"`
			Name  string      `json:"name"`
			Value interface{} `json:"value"`
		} `json:"parameters,omitempty"`
		Causes []struct {
			Class            string `json:"_class"`
			ShortDescription string `json:"shortDescription"`
			UpstreamBuild    int    `json:"upstreamBuild"`
			UpstreamProject  string `json:"upstreamProject"`
			UpstreamUrl      string `json:"upstreamUrl"`
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

type JenkinsListView struct {
	Class       string `json:"_class"`
	Description string `json:"description"`
	Jobs        []struct {
		Class string `json:"_class"`
		Name  string `json:"name"`
		URL   string `json:"url"`
		Color string `json:"color"`
	} `json:"jobs"`
	Name     string        `json:"name"`
	Property []interface{} `json:"property"`
	URL      string        `json:"url"`
}

func GetListViewJson(listViewUrl, username, password string) (*JenkinsListView, error) {
	listViewUrl, err := url.JoinPath(listViewUrl, "/api/json")
	if err != nil {
		return nil, err
	}

	var listView JenkinsListView
	err = GetAndUnmarshalUrl(listViewUrl, username, password, &listView)
	if err != nil {
		return nil, err
	}

	return &listView, nil
}

func GetBuildJson(buildURL, username, password string) (*JenkinsBuild, error) {
	buildURL, err := url.JoinPath(buildURL, "/api/json")
	if err != nil {
		return nil, err
	}

	var build JenkinsBuild
	err = GetAndUnmarshalUrl(buildURL, username, password, &build)
	if err != nil {
		return nil, err
	}

	return &build, nil
}

func GetAndUnmarshalUrl(jenkinsUrl, username, password string, unmarshalTo interface{}) error {
	log.Printf("fetching json from %v", jenkinsUrl)

	client := &http.Client{
		Timeout: time.Second * 20,
	}
	req, err := http.NewRequest("GET", jenkinsUrl, nil)
	if err != nil {
		return err
	}

	if username != "" && password != "" {
		// set up basic auth
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, unmarshalTo); err != nil {
		return err
	}

	return nil
}

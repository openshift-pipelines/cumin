package shared

import (
	"crypto/tls"
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

func GetListViewJson(listViewUrl string, insecure bool, username, password string) (*JenkinsListView, error) {
	listViewUrl, err := url.JoinPath(listViewUrl, "/api/json")
	if err != nil {
		return nil, err
	}

	var listView JenkinsListView
	_, err = GetAndUnmarshalUrl(listViewUrl, insecure, username, password, &listView)
	if err != nil {
		return nil, err
	}

	return &listView, nil
}

// GetBuildJson returns Build, status code, error
func GetBuildJson(buildURL string, insecure bool, username, password string) (*JenkinsBuild, int, error) {
	buildURL, err := url.JoinPath(buildURL, "/api/json")
	if err != nil {
		return nil, 0, err
	}

	var build JenkinsBuild
	statusCode, err := GetAndUnmarshalUrl(buildURL, insecure, username, password, &build)
	if err != nil {
		return nil, statusCode, err
	}

	return &build, statusCode, nil
}

func GetAndUnmarshalUrl(jenkinsUrl string, insecure bool, username, password string, unmarshalTo interface{}) (int, error) {
	log.Printf("fetching json from %v, insecureSkipVerify:%v", jenkinsUrl, insecure)

	client := &http.Client{
		Timeout: time.Second * 20,
	}

	if insecure {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client.Transport = customTransport
	}

	req, err := http.NewRequest("GET", jenkinsUrl, nil)
	if err != nil {
		return 0, err
	}

	if username != "" && password != "" {
		// set up basic auth
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if !IsStatusCodeOK(resp.StatusCode) {
		log.Printf("non-OK status code (%v) for url: %v", resp.StatusCode, jenkinsUrl)
		return resp.StatusCode, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if err := json.Unmarshal(body, unmarshalTo); err != nil {
		return resp.StatusCode, err
	}

	return resp.StatusCode, nil
}

func IsStatusCodeOK(statusCode int) bool {
	if statusCode < 200 || statusCode >= 300 {
		return false
	}
	return true
}

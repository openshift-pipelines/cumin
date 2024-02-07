package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/concaf/cumin/pkg/shared"
	"github.com/xanzy/go-gitlab"
)

type GitLabMergeRequest struct {
	Id    int
	URL   string
	Title string
}

type EndToEndFlow struct {
	GitlabMR              *GitLabMergeRequest
	CheckMerged           *JenkinsBuildView
	BuildPipeline         *JenkinsBuildView
	BuildPipelineChildren []*JenkinsBuildView
	ReleasePipeline       *JenkinsBuildView
	CVPPipeline           *JenkinsBuildView
	IndexImages           []string
}

type JenkinsBuildView struct {
	Name        string
	Number      int
	Logs        string
	Result      string
	Running     bool
	Duration    int
	Cause       string
	Url         string
	Previous    string
	DisplayName string
	Timestamp   int64
	Extra       interface{}
}

func GenerateEndToEndFlow(username, password, baseUrl, cvpUrl, checkMergedUrl string, insecure bool, generateNum int) ([]EndToEndFlow, error) {
	var e2eFlows []EndToEndFlow
	for i := 1; i <= generateNum; i++ {
		e2eFlow, err := GenerateEndToEndFlowSingle(username, password, baseUrl, cvpUrl, checkMergedUrl, insecure)
		if err != nil {
			return nil, err
		}
		e2eFlows = append(e2eFlows, *e2eFlow)
		if e2eFlow.CheckMerged.Previous == "" {
			break
		}
		checkMergedUrl = e2eFlow.CheckMerged.Previous
	}
	return e2eFlows, nil
}

// GenerateEndToEndFlowSingle
// baseUrl is something like https://<jenkins url>/job/pipelines-1.8-rhel-8/
func GenerateEndToEndFlowSingle(username, password, baseUrl, cvpUrl, checkMergedUrl string, insecure bool) (*EndToEndFlow, error) {
	checkMerged, err := GetBuildView(checkMergedUrl, insecure, username, password)
	if err != nil {
		return nil, err
	}
	e2eFlow := EndToEndFlow{
		CheckMerged: checkMerged,
	}

	glMr, err := getMRFromCheckMerged(username, password, checkMergedUrl, insecure)
	if err != nil {
		return nil, err
	}

	log.Printf("gitlab mr data: %v", glMr)
	e2eFlow.GitlabMR = glMr

	if checkMerged.Result != "success" {
		log.Printf("check-merged %v did not succeed, skipping further...", checkMerged.Url)
		return &e2eFlow, nil
	}

	// let's find a matching build-pipeline now
	latestBuildPipelineUrl, err := url.JoinPath(baseUrl, "job/build-pipeline/lastBuild")
	if err != nil {
		return nil, err
	}

	buildPipelineJson, statusCode, err := shared.GetBuildJson(latestBuildPipelineUrl, insecure, username, password)
	if !shared.IsStatusCodeOK(statusCode) {
		log.Printf("non-OK status code (%v) received for: %v", statusCode, latestBuildPipelineUrl)
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	matchingBuildPipelineUrl, err := matchBuildWithParent(username, password, "check-merged", checkMerged.Number, buildPipelineJson, insecure)
	if err != nil {
		return nil, err
	}
	if matchingBuildPipelineUrl == "" {
		log.Printf("no build-pipeline found for check-merged: %v", checkMergedUrl)
		return &e2eFlow, nil
	}

	buildPipeline, err := GetBuildView(matchingBuildPipelineUrl, insecure, username, password)
	if err != nil {
		return nil, err
	}
	e2eFlow.BuildPipeline = buildPipeline

	// let's move on to the build-pipeline children now
	buildPipelineJobsViewUrl, err := url.JoinPath(baseUrl, "job/openshift-pipelines/view/Build%20Jobs/")
	if err != nil {
		return nil, err
	}

	childJobs, err := getChildJobUrls(username, password, buildPipelineJobsViewUrl, insecure)
	if err != nil {
		return nil, err
	}
	if len(childJobs) < 1 {
		log.Printf("no child jobs found for build-pipeline")
		return &e2eFlow, nil
	}

	var operatorBundleBuild *shared.JenkinsBuild
	for _, job := range childJobs {
		log.Printf("parsing child job: %v", job)
		lastJobBuild, err := url.JoinPath(job, "lastBuild")
		if err != nil {
			return nil, err
		}

		jobBuildJson, statusCode, err := shared.GetBuildJson(lastJobBuild, insecure, username, password)
		if err != nil {
			return nil, err
		}
		if !shared.IsStatusCodeOK(statusCode) {
			log.Printf("non-OK status code (%v) received for: %v, ignoring...", statusCode, latestBuildPipelineUrl)
			continue
		}

		jobUrl, err := matchBuildWithParent(username, password, "build-pipeline", buildPipeline.Number, jobBuildJson, insecure)
		if err != nil {
			return nil, err
		}
		if jobUrl != "" {
			if strings.Contains(jobUrl, "job/openshift-pipelines-operator-bundle") {
				operatorBundleBuild, statusCode, err = shared.GetBuildJson(jobUrl, insecure, username, password)
				if err != nil {
					return nil, err
				}
				// this should not happen since job url should have been validated by now, error out
				if !shared.IsStatusCodeOK(statusCode) {
					return nil, fmt.Errorf("non-OK status code (%v) received for: %v", statusCode, jobUrl)
				}
			}
			jobBuildView, err := GetBuildView(jobUrl, insecure, username, password)
			if err != nil {
				return nil, err
			}
			e2eFlow.BuildPipelineChildren = append(e2eFlow.BuildPipelineChildren, jobBuildView)
		}
	}

	// let's match release-pipeline now
	releasePipelineUrl, err := url.JoinPath(baseUrl, "/job/release-pipeline/lastBuild")
	if err != nil {
		return nil, err
	}
	releasePipelineJson, statusCode, err := shared.GetBuildJson(releasePipelineUrl, insecure, username, password)
	if err != nil {
		return nil, err
	}
	if !shared.IsStatusCodeOK(statusCode) {
		log.Printf("non-OK status code (%v) for release pipeline url: %v, ignoring...", statusCode, releasePipelineUrl)
		return nil, nil
	}
	matchedReleasePipeline, err := matchBuildWithParent(username, password, "build-pipeline", buildPipeline.Number, releasePipelineJson, insecure)
	if err != nil {
		return nil, err
	}
	if matchedReleasePipeline != "" {
		releasePipelineView, err := GetBuildView(matchedReleasePipeline, insecure, username, password)
		if err != nil {
			return nil, err
		}
		e2eFlow.ReleasePipeline = releasePipelineView
	}

	// let's match cvp build now
	cvpJobUrl, err := url.JoinPath(cvpUrl, "lastBuild")
	if err != nil {
		return nil, err
	}

	if operatorBundleBuild == nil {
		log.Printf("no operator-bundle build found for build-pipeline (%v)", buildPipeline.Url)
		return &e2eFlow, nil
	}
	if strings.ToLower(operatorBundleBuild.Result) != "success" {
		log.Printf("operator-bundle build (%v) found for build-pipeline (%v) failed (%v)", operatorBundleBuild.URL, buildPipeline.Url, operatorBundleBuild.Result)
		return &e2eFlow, nil
	}
	matchingCVPBuild, err := findCVPFromBundleJob(cvpJobUrl, insecure, operatorBundleBuild)
	if err != nil {
		return nil, err
	}
	if matchingCVPBuild != "" {
		cvpBuildView, err := GetBuildView(matchingCVPBuild, insecure, "", "")
		if err != nil {
			return nil, err
		}
		e2eFlow.CVPPipeline = cvpBuildView

		if cvpBuildView.Result == "success" {
			indexImages, err := GetIndexImagesFromUrl(matchingCVPBuild)
			if err != nil {
				return nil, err
			}
			e2eFlow.IndexImages = indexImages
		}
	} else {
		log.Printf("could not find a matching CVP build for %v", operatorBundleBuild.URL)
	}

	return &e2eFlow, nil
}

func GetIndexImagesFromUrl(cvpJobUrl string) ([]string, error) {
	indexImagesUrl, err := url.JoinPath(cvpJobUrl, "artifact/index_images.yml")
	if err != nil {
		return nil, err
	}
	log.Printf("looking up index images at %v", indexImagesUrl)
	resp, err := http.Get(indexImagesUrl)
	if err != nil {
		// don't fail here
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected 200 OK for index builds, got %v", resp.StatusCode)
	}
	indexImagesBytes, err := io.ReadAll(resp.Body)
	log.Printf("index images: %v", string(indexImagesBytes))
	if err != nil {
		return nil, err
	}
	indexImages := strings.Split(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(indexImagesBytes), "[", ""), "]", ""), "'", ""), ",", ""), "\n", ""), "  ", " "), " ")

	return indexImages, nil
}

func getMRFromCheckMerged(username, password, checkMergedUrl string, insecure bool) (*GitLabMergeRequest, error) {
	checkMergedJson, statusCode, err := shared.GetBuildJson(checkMergedUrl, insecure, username, password)
	if err != nil {
		return nil, err
	}
	if !shared.IsStatusCodeOK(statusCode) {
		log.Printf("non-OK status code (%v) for check merged url: %v", statusCode, checkMergedUrl)
	}

	var mrIdString, repoUrl string
	for _, action := range checkMergedJson.Actions {
		for _, param := range action.Parameters {
			if param.Name == "gitlabMergeRequestIid" {
				mrIdString = param.Value.(string)
			}
			if param.Name == "gitlabTargetRepoHttpUrl" {
				// we don't want the trailing .git in the url
				repoUrl = strings.TrimSuffix(param.Value.(string), ".git")
			}
		}
	}
	if mrIdString == "" || repoUrl == "" {
		log.Printf("could not find mr id (%v) or repo url (%v)", mrIdString, repoUrl)
		return nil, nil
	}

	glMr := GitLabMergeRequest{}
	mrIdInt, err := strconv.Atoi(mrIdString)
	if err != nil {
		return nil, err
	}
	glMr.Id = mrIdInt
	// e.g. https://gitlab.cee.redhat.com/tekton/p12n/-/merge_requests/361
	mrUrl, err := url.JoinPath(repoUrl, "/-/merge_requests/", mrIdString)
	if err != nil {
		return nil, err
	}
	glMr.URL = mrUrl

	parsedUrl, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}

	gitlabClient, err := gitlab.NewClient("", gitlab.WithBaseURL(fmt.Sprintf("%v://%v", parsedUrl.Scheme, parsedUrl.Host)))
	if err != nil {
		return nil, err
	}
	glMrObj, _, err := gitlabClient.MergeRequests.GetMergeRequest(strings.Trim(parsedUrl.Path, "/"), mrIdInt, nil)
	if err != nil {
		return nil, err
	}
	glMr.Title = glMrObj.Title

	return &glMr, nil
}

// findCVPFromBundleJob we don't need creds to talk to cvp
// cvpUrl := https://<jenkins url>/view/all/job/cvp-redhat-operator-bundle-image-validation-test/
func findCVPFromBundleJob(cvpJobUrl string, insecure bool, operatorBundleBuild *shared.JenkinsBuild) (string, error) {
	type brewBuildNVR struct {
		NVR string `json:"nvr"`
	}

	log.Printf("will cvp pipeline (%v) match operator-bundle-build %v ?", cvpJobUrl, operatorBundleBuild.FullDisplayName)
	var nvr string
	for _, action := range operatorBundleBuild.Actions {
		if strings.Contains(action.Class, "ParametersAction") {
			for _, param := range action.Parameters {
				if param.Name == "brew_build_info" {
					brewNVR := brewBuildNVR{}
					brewBuildRawJson := strings.ReplaceAll(param.Value.(string), "\\", "")
					err := json.Unmarshal([]byte(brewBuildRawJson), &brewNVR)
					if err != nil {
						return "", err
					}
					nvr = brewNVR.NVR
					break
				}
			}
			if nvr != "" {
				break
			}
		}
	}
	if nvr == "" {
		return "", fmt.Errorf("could not find nvr in brew_build_info parameter in build %v", operatorBundleBuild.URL)
	}

	cvpBuildJson, statusCode, err := shared.GetBuildJson(cvpJobUrl, insecure, "", "")
	if err != nil {
		return "", err
	}
	if !shared.IsStatusCodeOK(statusCode) {
		log.Printf("non-OK status code (%v) for cvp url: %v, ignoring...", statusCode, cvpJobUrl)
		return "", nil
	}

	cvpBuildFound := false
	for _, action := range cvpBuildJson.Actions {
		if strings.Contains(action.Class, "ParametersAction") {
			for _, param := range action.Parameters {
				if param.Name == "CVP_PRODUCT_BREW_NVR" {
					if nvr == param.Value.(string) {
						cvpBuildFound = true
						break
					}
				}
			}
			if cvpBuildFound {
				break
			}
		}
	}

	if !cvpBuildFound {
		if cvpBuildJson.PreviousBuild.Number > 0 && cvpBuildJson.PreviousBuild.URL != "" {
			return findCVPFromBundleJob(cvpBuildJson.PreviousBuild.URL, insecure, operatorBundleBuild)
		} else {
			log.Printf("could not find matching cvp build for %v", operatorBundleBuild.FullDisplayName)
			return "", nil
		}
	}
	return cvpBuildJson.URL, nil
}

func matchBuildWithParent(username, password, parentJobName string, parentBuildNumber int, childBuild *shared.JenkinsBuild, insecure bool) (string, error) {
	childBuildFound := false
	for _, action := range childBuild.Actions {
		for _, cause := range action.Causes {
			if strings.Contains(cause.UpstreamProject, parentJobName) && strings.Contains(cause.UpstreamUrl, parentJobName) {
				if parentBuildNumber == cause.UpstreamBuild {
					log.Printf("found child build(%v) %v matching parent build (%v) %v", childBuild.FullDisplayName, childBuild.Number, parentJobName, parentBuildNumber)
					childBuildFound = true
					break
				} else {
					log.Printf("child build (%v) %v does not match parent build (%v) %v, has upstreamBuild %v", childBuild.FullDisplayName, childBuild.Number, parentJobName, parentBuildNumber, cause.UpstreamBuild)
				}
			}
		}
		if childBuildFound {
			break
		}
	}
	if !childBuildFound {
		// if the previous build exists, try to match it
		if childBuild.PreviousBuild.Number > 0 && childBuild.PreviousBuild.URL != "" {
			previousBuildPipelineJson, statusCode, err := shared.GetBuildJson(childBuild.PreviousBuild.URL, insecure, username, password)
			if err != nil {
				return "", err
			}
			if !shared.IsStatusCodeOK(statusCode) {
				return "", fmt.Errorf("non-OK status code (%v) for build: %v", statusCode, childBuild.PreviousBuild.URL)
			}
			return matchBuildWithParent(username, password, parentJobName, parentBuildNumber, previousBuildPipelineJson, insecure)
		} else {
			log.Printf("no matching child build (%v) found for parent build (%v) %v", childBuild.FullDisplayName, parentJobName, parentBuildNumber)
			return "", nil
		}
	}
	return childBuild.URL, nil
}

func getChildJobUrls(username, password, listViewUrl string, insecure bool) ([]string, error) {
	listView, err := shared.GetListViewJson(listViewUrl, insecure, username, password)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, job := range listView.Jobs {
		urls = append(urls, job.URL)
	}
	return urls, nil
}

func GetBuildView(jobUrl string, insecure bool, username, password string) (*JenkinsBuildView, error) {
	parsedUrl, err := url.Parse(jobUrl)
	if err != nil {
		return nil, err
	}

	build, statusCode, err := shared.GetBuildJson(jobUrl, insecure, username, password)
	if err != nil {
		return nil, err
	}
	if !shared.IsStatusCodeOK(statusCode) {
		return nil, fmt.Errorf("build view: non-OK status code(%v) for url: %v", statusCode, jobUrl)
	}

	splitUrl := strings.Split(parsedUrl.Path, "/job/")
	nameAndJob := splitUrl[len(splitUrl)-1]

	buildView := JenkinsBuildView{
		Name:        nameAndJob,
		Number:      build.Number,
		Logs:        filepath.Join(build.URL, "consoleFull"),
		Result:      strings.ToLower(build.Result),
		Running:     build.Building,
		Duration:    (build.Duration / 1000) / 60,
		Url:         build.URL,
		DisplayName: build.FullDisplayName,
		Timestamp:   build.Timestamp,
	}

	var buildCause string
	for _, action := range build.Actions {
		for _, cause := range action.Causes {
			buildCause = fmt.Sprintf("%v, %v", buildCause, cause.ShortDescription)
		}
	}
	buildView.Cause = buildCause

	if build.PreviousBuild.Number > 0 {
		buildView.Previous = build.PreviousBuild.URL
	}

	return &buildView, nil
}

func GenerateBuildViews(jobUrl string, insecure bool, username, password string, generateNum int) ([]*JenkinsBuildView, error) {
	var buildViews []*JenkinsBuildView

	for i := 1; i <= generateNum; i++ {
		currentBuildView, err := GetBuildView(jobUrl, insecure, username, password)
		if err != nil {
			return nil, err
		}
		buildViews = append(buildViews, currentBuildView)

		if currentBuildView.Previous == "" {
			break
		}
		jobUrl = currentBuildView.Previous
	}

	return buildViews, nil
}

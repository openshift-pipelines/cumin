package dashboard

import (
	"encoding/json"
	"fmt"
	"github.com/concaf/cumin/pkg/shared"
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

type EndToEndFlow struct {
	CheckMerged           *JenkinsBuildView
	BuildPipeline         *JenkinsBuildView
	BuildPipelineChildren []*JenkinsBuildView
	ReleasePipeline       *JenkinsBuildView
	CVPPipeline           *JenkinsBuildView
}

type JenkinsBuildView struct {
	Name     string
	Number   int
	Logs     string
	Result   string
	Running  bool
	Duration int
	Cause    string
	Url      string
	Previous string
}

// GenerateEndToEndFlow
// baseUrl is something like https://<jenkins url>/job/pipelines-1.8-rhel-8/
func GenerateEndToEndFlow(username, password, baseUrl, cvpUrl, checkMergedNum string) (*EndToEndFlow, error) {
	checkMergedUrl, err := url.JoinPath(baseUrl, "job/openshift-pipelines/job/check-merged/", checkMergedNum)
	if err != nil {
		return nil, err
	}

	checkMerged, err := GetBuildView(checkMergedUrl, username, password)
	if err != nil {
		return nil, err
	}
	e2eFlow := EndToEndFlow{
		CheckMerged: checkMerged,
	}

	if checkMerged.Result != "success" {
		log.Printf("check-merged %v did not succeed, skipping further...", checkMerged.Url)
		return &e2eFlow, nil
	}

	// let's find a matching build-pipeline now
	latestBuildPipelineUrl, err := url.JoinPath(baseUrl, "job/build-pipeline/lastBuild")
	if err != nil {
		return nil, err
	}

	buildPipelineJson, err := shared.GetBuildJson(latestBuildPipelineUrl, username, password)
	if err != nil {
		return nil, err
	}

	matchingBuildPipelineUrl, err := matchBuildWithParent(username, password, "check-merged", checkMerged.Number, buildPipelineJson)
	if err != nil {
		return nil, err
	}
	if matchingBuildPipelineUrl == "" {
		log.Printf("no build-pipeline found for check-merged: %v", checkMergedUrl)
		return &e2eFlow, nil
	}

	buildPipeline, err := GetBuildView(matchingBuildPipelineUrl, username, password)
	if err != nil {
		return nil, err
	}
	e2eFlow.BuildPipeline = buildPipeline

	// let's move on to the build-pipeline children now
	buildPipelineJobsViewUrl, err := url.JoinPath(baseUrl, "job/openshift-pipelines/view/Build%20Jobs/")
	if err != nil {
		return nil, err
	}

	childJobs, err := getChildJobUrls(username, password, buildPipelineJobsViewUrl)
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

		jobBuildJson, err := shared.GetBuildJson(lastJobBuild, username, password)
		if err != nil {
			return nil, err
		}

		jobUrl, err := matchBuildWithParent(username, password, "build-pipeline", buildPipeline.Number, jobBuildJson)
		if err != nil {
			return nil, err
		}
		if jobUrl != "" {
			if strings.Contains(jobUrl, "job/openshift-pipelines-operator-bundle") {
				operatorBundleBuild = jobBuildJson
			}
			jobBuildView, err := GetBuildView(jobUrl, username, password)
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
	releasePipelineJson, err := shared.GetBuildJson(releasePipelineUrl, username, password)
	if err != nil {
		return nil, err
	}
	matchedReleasePipeline, err := matchBuildWithParent(username, password, "build-pipeline", buildPipeline.Number, releasePipelineJson)
	if err != nil {
		return nil, err
	}
	if matchedReleasePipeline != "" {
		releasePipelineView, err := GetBuildView(matchedReleasePipeline, username, password)
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

	matchingCVPBuild, err := findCVPFromBundleJob(cvpJobUrl, operatorBundleBuild)
	if err != nil {
		return nil, err
	}
	if matchingCVPBuild != "" {
		cvpBuildView, err := GetBuildView(matchingCVPBuild, "", "")
		if err != nil {
			return nil, err
		}
		e2eFlow.CVPPipeline = cvpBuildView
	} else {
		log.Printf("could not find a matching CVP build for %v", operatorBundleBuild.URL)
	}

	return &e2eFlow, nil
}

// findCVPFromBundleJob we don't need creds to talk to cvp
// cvpUrl := https://<jenkins url>/view/all/job/cvp-redhat-operator-bundle-image-validation-test/
func findCVPFromBundleJob(cvpJobUrl string, operatorBundleBuild *shared.JenkinsBuild) (string, error) {
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

	cvpBuildJson, err := shared.GetBuildJson(cvpJobUrl, "", "")
	if err != nil {
		return "", err
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
			return findCVPFromBundleJob(cvpBuildJson.PreviousBuild.URL, operatorBundleBuild)
		} else {
			log.Printf("could not find matching cvp build for %v", operatorBundleBuild.FullDisplayName)
			return "", nil
		}
	}
	return cvpBuildJson.URL, nil
}

func matchBuildWithParent(username, password, parentJobName string, parentBuildNumber int, childBuild *shared.JenkinsBuild) (string, error) {
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
			previousBuildPipelineJson, err := shared.GetBuildJson(childBuild.PreviousBuild.URL, username, password)
			if err != nil {
				return "", err
			}
			return matchBuildWithParent(username, password, parentJobName, parentBuildNumber, previousBuildPipelineJson)
		} else {
			log.Printf("no matching child build (%v) found for parent build (%v) %v", childBuild.FullDisplayName, parentJobName, parentBuildNumber)
			return "", nil
		}
	}
	return childBuild.URL, nil
}

func getChildJobUrls(username, password, listViewUrl string) ([]string, error) {
	listView, err := shared.GetListViewJson(listViewUrl, username, password)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, job := range listView.Jobs {
		urls = append(urls, job.URL)
	}
	return urls, nil
}

func GetBuildView(jobUrl, username, password string) (*JenkinsBuildView, error) {
	parsedUrl, err := url.Parse(jobUrl)
	if err != nil {
		return nil, err
	}

	build, err := shared.GetBuildJson(jobUrl, username, password)
	if err != nil {
		return nil, err
	}

	splitUrl := strings.Split(parsedUrl.Path, "/job/")
	nameAndJob := splitUrl[len(splitUrl)-1]

	buildView := JenkinsBuildView{
		Name:     nameAndJob,
		Number:   build.Number,
		Logs:     filepath.Join(build.URL, "consoleFull"),
		Result:   strings.ToLower(build.Result),
		Running:  build.Building,
		Duration: (build.Duration / 1000) / 60,
		Url:      build.URL,
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

func GenerateBuildViews(url, username, password string) ([]JenkinsBuildView, error) {
	var buildViews []JenkinsBuildView
	currentBuildView, err := GetBuildView(url, username, password)
	if err != nil {
		return nil, err
	}
	buildViews = append(buildViews, *currentBuildView)

	if currentBuildView.Previous != "" {
		previousBuildView, err := GetBuildView(currentBuildView.Previous, username, password)
		if err != nil {
			return nil, err
		}
		buildViews = append(buildViews, *previousBuildView)
	}

	return buildViews, err
}

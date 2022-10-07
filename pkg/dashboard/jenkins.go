package dashboard

import (
	"fmt"
	"github.com/concaf/cumin/pkg/shared"
	"path/filepath"
	"strings"
)

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

func GetBuildView(url, username, password string) (*JenkinsBuildView, error) {
	build, err := shared.GetBuildJSONFromURL(url, username, password)
	if err != nil {
		return nil, err
	}

	buildView := JenkinsBuildView{
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
	currentBuildView.Name = "current build"
	buildViews = append(buildViews, *currentBuildView)

	if currentBuildView.Previous != "" {
		previousBuildView, err := GetBuildView(currentBuildView.Previous, username, password)
		if err != nil {
			return nil, err
		}
		previousBuildView.Name = "previous build"
		buildViews = append(buildViews, *previousBuildView)
	}

	return buildViews, err
}

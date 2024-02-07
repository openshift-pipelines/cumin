package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/concaf/cumin/pkg/dashboard"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	buildURL                string
	insecureBuildURL        bool
	viewGenerateNum         int
	viewLastSuccessfulBuild bool
	getIndexImages          bool
)

// buildViewCmd represents the jenkins command
var buildViewCmd = &cobra.Command{
	Use:   "build-view",
	Short: "jenkins build view",
	Run: func(cmd *cobra.Command, args []string) {
		buildViews, err := dashboard.GenerateBuildViews(buildURL, insecureBuildURL, jenkinsUsername, jenkinsPassword, viewGenerateNum)
		if err != nil {
			log.Fatal(err)
		}

		successfulPresent := false
		for _, build := range buildViews {
			if build.Result == "success" {
				successfulPresent = true
				break
			}
		}
		// we don't want lastSuccessfulBuild if there is already one inside there
		if viewLastSuccessfulBuild && !successfulPresent {
			// https://<jenkins url>/job/pipelines-1.8-rhel-8/job/build-pipeline/41/
			splitUrl := strings.Split(strings.Trim(buildURL, "/"), "/")
			splitUrl = splitUrl[:len(splitUrl)-1]
			splitUrl = append(splitUrl, "lastSuccessfulBuild")
			lastSuccessfulUrl := strings.Join(splitUrl, "/")

			lastSuccessfulView, err := dashboard.GetBuildView(lastSuccessfulUrl, insecureBuildURL, jenkinsUsername, jenkinsPassword)
			if err != nil {
				log.Fatal(err)
			}
			buildViews = append(buildViews, lastSuccessfulView)
		}

		if getIndexImages {
			for _, view := range buildViews {
				if view.Result != "success" {
					log.Printf("skipping getting index images for %v since the build result is %v", view.Name, view.Result)
					continue
				}
				indexImages, err := dashboard.GetIndexImagesFromUrl(view.Url)
				if err != nil {
					log.Fatal(err)
				}
				view.Extra = indexImages
			}
		}

		log.Println("marshalling now...")
		buildViewsYaml, err := yaml.Marshal(buildViews)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(buildViewsYaml))
	},
}

func init() {
	generateCmd.AddCommand(buildViewCmd)

	buildViewCmd.Flags().StringVar(&buildURL, "build", "", "url for the jenkins build")
	buildViewCmd.MarkFlagRequired("build")
	buildViewCmd.Flags().BoolVar(&insecureBuildURL, "insecure", true, "insecure skip verify")
	buildViewCmd.Flags().IntVar(&viewGenerateNum, "generate-num", 3, "number of views to generate")
	buildViewCmd.Flags().BoolVar(&viewLastSuccessfulBuild, "last-successful-build", true, "do you want to generate the last successful build as well?")
	buildViewCmd.Flags().BoolVar(&getIndexImages, "index-images", false, "get index images for cvp job urls?")
}

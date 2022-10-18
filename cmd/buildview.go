package cmd

import (
	"fmt"
	"github.com/concaf/cumin/pkg/dashboard"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
)

var (
	buildURL string
)

// buildViewCmd represents the jenkins command
var buildViewCmd = &cobra.Command{
	Use:   "build-view",
	Short: "jenkins build view",
	Run: func(cmd *cobra.Command, args []string) {
		buildViews, err := dashboard.GenerateBuildViews(buildURL, jenkinsUsername, jenkinsPassword)
		if err != nil {
			log.Fatal(err)
		}
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
}

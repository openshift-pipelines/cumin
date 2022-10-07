package cmd

import (
	"fmt"
	"github.com/concaf/cumin/pkg/dashboard"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var (
	buildURL        string
	jenkinsUsername string
	jenkinsPassword string
)

// jenkinsCmd represents the jenkins command
var jenkinsCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "A brief description of your command",
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
	generateCmd.AddCommand(jenkinsCmd)

	jenkinsCmd.Flags().StringVar(&buildURL, "build", "", "url for the jenkins build")
	jenkinsCmd.MarkFlagRequired("build")

	jenkinsCmd.Flags().StringVarP(&jenkinsUsername, "username", "u", os.Getenv("JENKINS_USERNAME"), "jenkins username")
	jenkinsCmd.Flags().StringVarP(&jenkinsPassword, "password", "p", os.Getenv("JENKINS_PASSWORD"), "jenkins password")
}

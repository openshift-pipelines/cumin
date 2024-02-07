/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"net/url"

	"github.com/concaf/cumin/pkg/dashboard"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var (
	baseUrl             string
	cvpUrl              string
	insecureFlowUrl     bool
	checkMergedNum      string
	generateNum         int
	lastSuccessfulBuild bool
)

// e2eflowCmd represents the e2eflow command
var e2eflowCmd = &cobra.Command{
	Use:   "e2eflow",
	Short: "e2eflow of a check-merged job",
	Run: func(cmd *cobra.Command, args []string) {
		checkMergedUrl, err := url.JoinPath(baseUrl, "job/openshift-pipelines/job/check-merged/", checkMergedNum)
		if err != nil {
			log.Fatal(err)
		}
		e2eFlow, err := dashboard.GenerateEndToEndFlow(jenkinsUsername, jenkinsPassword, baseUrl, cvpUrl, checkMergedUrl, insecureFlowUrl, generateNum)
		if err != nil {
			log.Fatal(err)
		}

		successfulPresent := false
		for _, build := range e2eFlow {
			if build.CheckMerged.Result == "success" {
				successfulPresent = true
				break
			}
		}
		// we don't want lastSuccessfulBuild if there is already one inside there
		if lastSuccessfulBuild && !successfulPresent {
			lastSuccessfulUrl, err := url.JoinPath(baseUrl, "job/openshift-pipelines/job/check-merged/lastSuccessfulBuild")
			if err != nil {
				log.Fatal(err)
			}
			lastSuccessfulFlow, err := dashboard.GenerateEndToEndFlowSingle(jenkinsUsername, jenkinsPassword, baseUrl, cvpUrl, lastSuccessfulUrl, insecureFlowUrl)
			if err != nil {
				log.Fatal(err)
			}
			e2eFlow = append(e2eFlow, *lastSuccessfulFlow)
		}

		e2eFlowYaml, err := yaml.Marshal(e2eFlow)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(e2eFlowYaml))
	},
}

func init() {
	generateCmd.AddCommand(e2eflowCmd)

	e2eflowCmd.Flags().StringVar(&baseUrl, "base", "", "base url of jenkins job, e.g. https://<jenkins url>/job/pipelines-1.8-rhel-8/")
	e2eflowCmd.MarkFlagRequired("base")

	e2eflowCmd.Flags().StringVar(&cvpUrl, "cvp", "", "url of cvp jenkins, e.g. https://<jenkins url>/view/all/job/cvp-redhat-operator-bundle-image-validation-test/")
	e2eflowCmd.MarkFlagRequired("cvp")

	e2eflowCmd.Flags().BoolVar(&insecureFlowUrl, "insecure", true, "insecure skip verify")

	e2eflowCmd.Flags().StringVar(&checkMergedNum, "check-merged", "lastBuild", "number of check-merged job")
	e2eflowCmd.Flags().IntVar(&generateNum, "generate-num", 3, "number of builds to generate")
	e2eflowCmd.Flags().BoolVar(&lastSuccessfulBuild, "last-successful-build", true, "do you want to generate the last successful build as well?")
}

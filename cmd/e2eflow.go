/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/concaf/cumin/pkg/dashboard"
	"gopkg.in/yaml.v3"
	"log"

	"github.com/spf13/cobra"
)

var (
	baseUrl        string
	cvpUrl         string
	checkMergedNum string
)

// e2eflowCmd represents the e2eflow command
var e2eflowCmd = &cobra.Command{
	Use:   "e2eflow",
	Short: "e2eflow of a check-merged job",
	Run: func(cmd *cobra.Command, args []string) {
		e2eFlow, err := dashboard.GenerateEndToEndFlow(jenkinsUsername, jenkinsPassword, baseUrl, cvpUrl, checkMergedNum)
		if err != nil {
			log.Fatal(err)
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

	e2eflowCmd.Flags().StringVar(&baseUrl, "cvp", "", "url of cvp jenkins")
	e2eflowCmd.Flags().StringVar(&checkMergedNum, "check-merged", "lastBuild", "number of check-merged job")
}

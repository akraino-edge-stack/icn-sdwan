// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update(Put) the resources from input file or url(without body) from command line",
	Run: func(cmd *cobra.Command, args []string) {
		var c RestyClient
		if len(token) > 0 {
			c = NewRestClientToken(token[0])
		} else {
			c = NewRestClient()
		}
		if len(inputFiles) > 0 {
			resources := readResources()
			for _, res := range resources {
				if res.file != "" {
					err := c.RestClientMultipartPut(res.anchor, res.body, res.file)
					if err != nil && err.Error() != "Server Error" {
						fmt.Println("Update: ", res.anchor, "Error: ", err)
					}
				} else if len(res.files) > 0 {
					err := c.RestClientMultipartPutMultipleFiles(res.anchor, res.body, res.files)
					if err != nil && err.Error() != "Server Error" {
						fmt.Println("Update: ", res.anchor, "Error: ", err)
					}
				} else {
					err := c.RestClientPut(res.anchor, res.body)
					if err != nil && err.Error() != "Server Error" {
						fmt.Println("Update: ", res.anchor, "Error: ", err)
					}
				}
			}
		} else {
			fmt.Println("Error: No args ")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringSliceVarP(&inputFiles, "filename", "f", []string{}, "Filename of the input file")
	updateCmd.Flags().StringSliceVarP(&valuesFiles, "values", "v", []string{}, "Template Values to go with the input template file")
	updateCmd.Flags().StringSliceVarP(&token, "token", "t", []string{}, "Token for EWO API")
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the resources from input file or url from command line",
	Run: func(cmd *cobra.Command, args []string) {
		var c RestyClient
		if len(token) > 0 {
			c = NewRestClientToken(token[0])
		} else {
			c = NewRestClient()
		}
		if len(inputFiles) > 0 {
			resources := readResources()
			for i := len(resources) - 1; i >= 0; i-- {
				res := resources[i]
				c.RestClientDelete(res.anchor, res.body)
			}
		} else if len(args) >= 1 {
			c.RestClientDeleteAnchor(args[0])
		} else {
			fmt.Println("Error: No args ")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringSliceVarP(&inputFiles, "filename", "f", []string{}, "Filename of the input file")
	deleteCmd.Flags().StringSliceVarP(&valuesFiles, "values", "v", []string{}, "Template Values to go with the input template file")
	deleteCmd.Flags().StringSliceVarP(&token, "token", "t", []string{}, "Token for EWO API")
}

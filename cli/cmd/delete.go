/*
Copyright 2019 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cortexlabs/cortex/pkg/lib/errors"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [APP_NAME]",
	Short: "delete a deployment",
	Long:  "Delete a deployment.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var appName string
		var err error
		if len(args) == 1 {
			appName = args[0]
		} else {
			appName, err = appNameFromConfig()
			if err != nil {
				errors.Exit(err)
			}
		}

		fmt.Println(appName)
	},
}

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
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	addDeploymentNameFlag(getCmd)
	addEnvFlag(getCmd)
	addWatchFlag(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get [RESOURCE_NAME]",
	Short: "get information about APIs",
	Long:  "Get information about APIs.",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		rerun(func() (string, error) {
			return runGet(cmd, args)
		})
	},
}

func runGet(cmd *cobra.Command, args []string) (string, error) {
	deploymentName, err := DeploymentNameFromFlagOrConfig()
	if err != nil {
		return "", err
	}

	return deploymentName + " " + strings.Join(args, ", "), nil
}

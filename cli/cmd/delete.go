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
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/cortexlabs/cortex/pkg/lib/files"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [APP_NAME]",
	Short: "delete a deployment",
	Long:  "Delete a deployment.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		cortexConfig := &CortexConfig{}
		cortexConfigPath := filepath.Join(dir, "cortex.json")
		bytes, err := files.ReadFileBytes(cortexConfigPath)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(bytes, cortexConfig)

		dockerClient, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		ctx := context.Background()
		for _, localDeployment := range cortexConfig.LocalDeployments {
			err := dockerClient.ContainerRemove(ctx, localDeployment.ContainerID, types.ContainerRemoveOptions{
				Force: true,
			})
			if err != nil {
				panic(err)
			}
		}
	},
}

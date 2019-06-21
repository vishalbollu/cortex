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
	"io"
	"os"
	"path/filepath"

	"github.com/cortexlabs/cortex/pkg/lib/debug"
	"github.com/cortexlabs/cortex/pkg/lib/files"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

func init() {
	addAppNameFlag(logsCmd)
	addEnvFlag(logsCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs RESOURCE_NAME",
	Short: "get logs for a resource",
	Long:  "Get logs for a resource.",
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
		var localDeployment LocalDeployment
		for _, ld := range cortexConfig.LocalDeployments {
			if ld.Name == args[0] {
				localDeployment = *ld
			}
		}
		debug.Pp(localDeployment)
		reader, err := dockerClient.ContainerLogs(ctx, localDeployment.ContainerID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if err != nil {
			panic(err)
		}
		defer reader.Close()
		_, err = io.Copy(os.Stdout, reader)
		if err != nil && err != io.EOF {
			panic(err)
		}
	},
}

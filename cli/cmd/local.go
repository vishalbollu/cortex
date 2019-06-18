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
	"fmt"

	"github.com/docker/docker/api/types"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
)

func init() {
	localCmd.PersistentFlags()
	addEnvFlag(localCmd)
}

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "local an application",
	Long:  "local an application.",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		s3_path := args[0]
		key := args[1]
		secret := args[2]
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		ctx := context.Background()
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: "cortexlabs/serving-tf",
			Env:   []string{"AWS_ACCESS_KEY_ID=" + key, "AWS_SECRET_ACCESS_KEY=" + secret, "EXTERNAL_MODEL_PATH=" + s3_path},
			ExposedPorts: nat.PortSet{
				"8888/tcp": struct{}{},
			},
		}, &container.HostConfig{
			PortBindings: nat.PortMap{
				"8888/tcp": []nat.PortBinding{
					{
						HostPort: "8888",
					},
				},
			},
		}, nil, "")
		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}

		fmt.Println("localhost:8888")
	},
}

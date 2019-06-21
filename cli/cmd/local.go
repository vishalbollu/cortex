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
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"

	"github.com/cortexlabs/cortex/pkg/lib/debug"
	"github.com/cortexlabs/cortex/pkg/lib/files"
)

type UserConfig struct {
	LocalAPIs []*LocalAPI `json:"local_apis"`
}

type LocalAPI struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type LocalDeployment struct {
	ContainerID string `json:"container_id"`
	Port        string `json:"port"`
	TFPort      string `json:"tf_port"`
	*LocalAPI
}

type CortexConfig struct {
	LocalDeployments []*LocalDeployment `json:"local_deployments"`
}

func init() {
	localCmd.PersistentFlags()
	addEnvFlag(localCmd)
}

func localMode() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(dir, "app.json")
	bytes, err := files.ReadFileBytes(appDir)
	if err != nil {
		panic(err)
	}

	userConfig := &UserConfig{}
	json.Unmarshal(bytes, userConfig)

	cortexConfig := &CortexConfig{}
	cortexConfigPath := filepath.Join(dir, "cortex.json")

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	port := 8000
	tfPort := 9000
	for id, localAPI := range userConfig.LocalAPIs {
		cleanFilePath, err := files.FullPath(localAPI.Path)
		if err != nil {
			panic(err)
		}

		dir, filename := filepath.Split(cleanFilePath)
		if len(filename) == 0 {
			dir, filename = filepath.Split(dir)
		}

		debug.Pp(dir)
		ctx := context.Background()
		mountBase := "/mnt/model"
		mountPath := filepath.Join(mountBase, filename)
		apiPort := port + id
		resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
			Image: "cortexlabs/serving-tf",
			Env:   []string{"EXTERNAL_MODEL_PATH=" + mountPath, "MODEL_BASE_PATH=" + mountPath},
			ExposedPorts: nat.PortSet{
				"8888/tcp": struct{}{},
				"9000/tcp": struct{}{},
			},
			Volumes: map[string]struct{}{
				mountBase: struct{}{},
			},
		}, &container.HostConfig{
			PortBindings: nat.PortMap{
				"8888/tcp": []nat.PortBinding{
					{
						HostPort: fmt.Sprintf("%d", apiPort),
					},
				},
				"9000/tcp": []nat.PortBinding{
					{
						HostPort: fmt.Sprintf("%d", tfPort),
					},
				},
			},
			Mounts: []mount.Mount{
				mount.Mount{
					Type:   mount.TypeBind,
					Source: dir,
					Target: mountBase,
				},
			},
		}, nil, "")
		if err != nil {
			panic(err)
		}

		if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}

		cortexConfig.LocalDeployments = append(cortexConfig.LocalDeployments, &LocalDeployment{
			ContainerID: resp.ID,
			Port:        fmt.Sprintf("%d", apiPort),
			TFPort:      fmt.Sprintf("%d", tfPort),
			LocalAPI:    localAPI,
		})

		fmt.Printf("%s http://localhost:%d\n", localAPI.Name, apiPort)
	}

	jsonBytes, err := json.Marshal(cortexConfig)
	if err != nil {
		panic(err)
	}
	files.WriteFile(cortexConfigPath, jsonBytes, 0644)
}

func kubeMode() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	appDir := filepath.Join(dir, "app.json")
	bytes, err := files.ReadFileBytes(appDir)
	if err != nil {
		panic(err)
	}

	userConfig := &UserConfig{}
	json.Unmarshal(bytes, userConfig)

}

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "local an application",
	Long:  "local an application.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		localMode()

		// s3_path := args[0]
		// key := args[1]
		// secret := args[2]
		// cli, err := client.NewEnvClient()
		// if err != nil {
		// 	panic(err)
		// }
		// ctx := context.Background()
		// resp, err := cli.ContainerCreate(ctx, &container.Config{
		// 	Image: "cortexlabs/serving-tf",
		// 	Env:   []string{"EXTERNAL_MODEL_PATH=" + s3_path},
		// 	ExposedPorts: nat.PortSet{
		// 		"8888/tcp": struct{}{},
		// 	},
		// }, &container.HostConfig{
		// 	PortBindings: nat.PortMap{
		// 		"8888/tcp": []nat.PortBinding{
		// 			{
		// 				HostPort: "8888",
		// 			},
		// 		},
		// 	},
		// }, nil, "")
		// if err != nil {
		// 	panic(err)
		// }

		// if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		// 	panic(err)
		// }

		// fmt.Println("localhost:8888")
	},
}

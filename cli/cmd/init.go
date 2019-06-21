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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cortexlabs/cortex/pkg/lib/errors"
	"github.com/cortexlabs/cortex/pkg/lib/files"
	"github.com/cortexlabs/cortex/pkg/lib/urls"
)

var initCmd = &cobra.Command{
	Use:   "init DEPLOYMENT_NAME",
	Short: "initialize a deployment",
	Long:  "Initialize a deployment.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deploymentName := args[0]

		if err := urls.CheckDNS1123(deploymentName); err != nil {
			errors.Exit(err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			errors.Exit(err)
		}

		if currentRoot := cortexRootOrBlank(); currentRoot != "" {
			errors.Exit(ErrorCLIAlreadyInCortexDir(currentRoot))
		}

		cortexRoot := filepath.Join(cwd, deploymentName)
		var createdFiles []string

		if _, err := files.CreateDirIfMissing(cortexRoot); err != nil {
			errors.Exit(err)
		}

		for path, content := range initFiles(deploymentName) {
			createdFiles = writeFile(path, content, cortexRoot, createdFiles)
		}

		fmt.Println("Created files:")
		fmt.Println(files.FileTree(createdFiles, cwd, files.DirsOnBottom))
	},
}

func writeFile(subPath string, content string, root string, createdFiles []string) []string {
	path := filepath.Join(root, subPath)
	if _, err := files.CreateDirIfMissing(filepath.Dir(path)); err != nil {
		errors.Exit(err)
	}
	err := files.WriteFile(path, []byte(content), 0664)
	if err != nil {
		errors.Exit(err)
	}
	return append(createdFiles, path)
}

func initFiles(deploymentName string) map[string]string {
	return map[string]string{
		"cortex.yaml": fmt.Sprintf("- kind: deployment\n  name: %s\n", deploymentName) + `
## Sample API:
#
# - kind: api
#   name: my-api
#   model: s3://my-bucket/my-model.zip
#   replicas: 1
`,

		"samples.json": `{
  "samples": [
    {
      "key1": "value1",
      "key2": "value2",
      "key3": "value3"
    }
  ]
}
`,
	}
}

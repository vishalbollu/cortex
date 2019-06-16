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
	Use:   "init APP_NAME",
	Short: "initialize an application",
	Long:  "Initialize an application.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		if err := urls.CheckDNS1123(appName); err != nil {
			errors.Exit(err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			errors.Exit(err)
		}

		if currentRoot := appRootOrBlank(); currentRoot != "" {
			errors.Exit(ErrorCLIAlreadyInAppDir(currentRoot))
		}

		appRoot := filepath.Join(cwd, appName)
		var createdFiles []string

		if _, err := files.CreateDirIfMissing(appRoot); err != nil {
			errors.Exit(err)
		}

		for path, content := range appInitFiles(appName) {
			createdFiles = writeFile(path, content, appRoot, createdFiles)
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

func appInitFiles(appName string) map[string]string {
	return map[string]string{
		"app.yaml": fmt.Sprintf("- kind: app\n  name: %s\n", appName),

		"resources/apis.yaml": `## Sample API:
#
# - kind: api
#   name: my-api
#   model: @my_model
#   compute:
#     replicas: 1
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

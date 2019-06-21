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
)

var predictPrintJSON bool

func init() {
	predictCmd.PersistentFlags().BoolVarP(&predictPrintJSON, "json", "j", false, "print the raw json response")
	addDeploymentNameFlag(predictCmd)
	addEnvFlag(predictCmd)
}

type PredictResponse struct {
	ResourceID  string       `json:"resource_id"`
	Predictions []Prediction `json:"predictions"`
}

type Prediction struct {
	Prediction         interface{} `json:"prediction"`
	PredictionReversed interface{} `json:"prediction_reversed"`
	TransformedSample  interface{} `json:"transformed_sample"`
	Response           interface{} `json:"response"`
}

var predictCmd = &cobra.Command{
	Use:   "predict API_NAME SAMPLES_FILE",
	Short: "make predictions",
	Long:  "Make predictions.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
	},
}

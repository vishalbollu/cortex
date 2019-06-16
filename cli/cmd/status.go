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
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cortexlabs/cortex/pkg/operator/api/resource"
	"github.com/cortexlabs/cortex/pkg/operator/api/schema"
)

func init() {
	addAppNameFlag(statusCmd)
	addEnvFlag(statusCmd)
	addWatchFlag(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "get resource statuses",
	Long:  "Get resource statuses.",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		rerun(func() (string, error) {
			return runStatus(cmd, args)
		})
	},
}

func runStatus(cmd *cobra.Command, args []string) (string, error) {
	return strings.Join(args, ","), nil
}

func resourceStatusesStr(resourcesRes *schema.GetResourcesResponse) string {
	var statuses = make([]resource.Status, len(resourcesRes.APIGroupStatuses))
	i := 0
	for _, apiGroupStatus := range resourcesRes.APIGroupStatuses {
		statuses[i] = apiGroupStatus
		i++
	}
	return "APIs:                  " + StatusStr(statuses)
}

func StatusStr(statuses []resource.Status) string {
	if len(statuses) == 0 {
		return "none"
	}

	messageBuckets := make(map[int][]string)
	for _, status := range statuses {
		bucketKey := status.GetCode().SortBucket()
		messageBuckets[bucketKey] = append(messageBuckets[bucketKey], status.Message())
	}

	var bucketKeys []int
	for bucketKey := range messageBuckets {
		bucketKeys = append(bucketKeys, bucketKey)
	}
	sort.Ints(bucketKeys)

	var messageItems []string

	for _, bucketKey := range bucketKeys {
		messageCounts := make(map[string]int)
		for _, message := range messageBuckets[bucketKey] {
			messageCounts[message]++
		}

		var messages []string
		for message := range messageCounts {
			messages = append(messages, message)
		}
		sort.Strings(messages)

		for _, message := range messages {
			messageItem := fmt.Sprintf("%d %s", messageCounts[message], message)
			messageItems = append(messageItems, messageItem)
		}
	}

	return strings.Join(messageItems, " | ")
}

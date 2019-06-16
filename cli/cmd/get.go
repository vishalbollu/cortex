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
	"time"

	"github.com/spf13/cobra"

	"github.com/cortexlabs/cortex/pkg/lib/sets/strset"
	s "github.com/cortexlabs/cortex/pkg/lib/strings"
	libtime "github.com/cortexlabs/cortex/pkg/lib/time"
	"github.com/cortexlabs/cortex/pkg/lib/urls"
	"github.com/cortexlabs/cortex/pkg/operator/api/context"
	"github.com/cortexlabs/cortex/pkg/operator/api/resource"
	"github.com/cortexlabs/cortex/pkg/operator/api/schema"
	"github.com/cortexlabs/cortex/pkg/operator/api/userconfig"
)

func init() {
	addAppNameFlag(getCmd)
	addEnvFlag(getCmd)
	addWatchFlag(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get [RESOURCE_NAME]",
	Short: "get information about resources",
	Long:  "Get information about resources.",
	Args:  cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		rerun(func() (string, error) {
			return runGet(cmd, args)
		})
	},
}

func runGet(cmd *cobra.Command, args []string) (string, error) {
	appName, err := AppNameFromFlagOrConfig()
	if err != nil {
		return "", err
	}

	return appName + " " + strings.Join(args, ", "), nil
}

func apisStr(apiGroupStatuses map[string]*resource.APIGroupStatus) string {
	if len(apiGroupStatuses) == 0 {
		return "None\n"
	}

	strings := make(map[string]string)
	for name, apiGroupStatus := range apiGroupStatuses {
		strings[name] = apiResourceRow(apiGroupStatus)
	}
	return apisHeader() + strMapToStr(strings)
}

func describeAPI(name string, resourcesRes *schema.GetResourcesResponse) (string, error) {
	groupStatus := resourcesRes.APIGroupStatuses[name]
	if groupStatus == nil {
		return "", userconfig.ErrorUndefinedResource(name, resource.APIType)
	}

	ctx := resourcesRes.Context
	api := ctx.APIs[name]

	var staleReplicas int32
	var ctxAPIStatus *resource.APIStatus
	var anyAPIStatus *resource.APIStatus
	for _, apiStatus := range resourcesRes.APIStatuses {
		if apiStatus.APIName != name {
			continue
		}
		anyAPIStatus = apiStatus
		if api != nil && apiStatus.ResourceID == api.ID {
			ctxAPIStatus = apiStatus
		}
		staleReplicas += apiStatus.TotalStaleReady()
	}

	out := titleStr("Summary")
	out += "Status:            " + groupStatus.Message() + "\n"
	if ctxAPIStatus != nil {
		out += fmt.Sprintf("Updated replicas:  %d/%d ready\n", ctxAPIStatus.ReadyUpdated, ctxAPIStatus.RequestedReplicas)
	}
	if staleReplicas != 0 {
		out += fmt.Sprintf("Stale replicas:    %d ready\n", staleReplicas)
	}
	out += "Created at:        " + libtime.LocalTimestamp(groupStatus.Start) + "\n"
	if groupStatus.ActiveStatus != nil && groupStatus.ActiveStatus.Start != nil {
		out += "Refreshed at:      " + libtime.LocalTimestamp(groupStatus.ActiveStatus.Start) + "\n"
	}

	out += titleStr("Endpoint")
	out += "URL:      " + urls.Join(resourcesRes.APIsBaseURL, anyAPIStatus.Path) + "\n"
	out += "Method:   POST\n"
	out += `Header:   "Content-Type: application/json"` + "\n"

	if api.Model != nil {
		model := ctx.Models[api.ModelName]
		resIDs := strset.New()
		combinedInput := []interface{}{model.Input, model.TrainingInput}
		for _, res := range ctx.ExtractCortexResources(combinedInput, resource.ConstantType, resource.RawColumnType, resource.AggregateType, resource.TransformedColumnType) {
			resIDs.Add(res.GetID())
			resIDs.Merge(ctx.AllComputedResourceDependencies(res.GetID()))
		}
		var samplePlaceholderFields []string
		for rawColumnName, rawColumn := range ctx.RawColumns {
			if resIDs.Has(rawColumn.GetID()) {
				fieldStr := fmt.Sprintf("\"%s\": %s", rawColumnName, rawColumn.GetColumnType().JSONPlaceholder())
				samplePlaceholderFields = append(samplePlaceholderFields, fieldStr)
			}
		}
		sort.Strings(samplePlaceholderFields)
		samplesPlaceholderStr := `{ "samples": [ { ` + strings.Join(samplePlaceholderFields, ", ") + " } ] }"
		out += "Payload:  " + samplesPlaceholderStr + "\n"
	}
	if api != nil {
		out += resourceStr(api.API)
	}

	return out, nil
}

func resourceStr(resource userconfig.Resource) string {
	return titleStr("Configuration") + s.Obj(resource) + "\n"
}

func strMapToStr(strings map[string]string) string {
	var keys []string
	for key := range strings {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	out := ""
	for _, key := range keys {
		out += strings[key] + "\n"
	}

	return out
}

func dataResourceRow(name string, resource context.Resource, dataStatuses map[string]*resource.DataStatus) string {
	dataStatus := dataStatuses[resource.GetID()]
	return resourceRow(name, dataStatus.Message(), dataStatus.End)
}

func apiResourceRow(groupStatus *resource.APIGroupStatus) string {
	var updatedAt *time.Time
	if groupStatus.ActiveStatus != nil {
		updatedAt = groupStatus.ActiveStatus.Start
	}
	return resourceRow(groupStatus.APIName, groupStatus.Message(), updatedAt)
}

func resourceRow(name string, status string, startTime *time.Time) string {
	if len(name) > 33 {
		name = name[0:30] + "..."
	}
	if len(status) > 23 {
		status = status[0:20] + "..."
	}
	timeSince := libtime.Since(startTime)
	return stringifyRow(name, status, timeSince)
}

func dataResourcesHeader() string {
	return stringifyRow("NAME", "STATUS", "AGE") + "\n"
}

func apisHeader() string {
	return stringifyRow("NAME", "STATUS", "LAST UPDATE") + "\n"
}

func stringifyRow(name string, status string, timeSince string) string {
	return fmt.Sprintf("%-35s%-24s%s", name, status, timeSince)
}

func titleStr(title string) string {
	titleLength := len(title)
	top := strings.Repeat("-", titleLength)
	bottom := strings.Repeat("-", titleLength)
	return "\n" + top + "\n" + title + "\n" + bottom + "\n\n"
}

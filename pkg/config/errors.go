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

package config

import (
	"fmt"
	"strings"

	"github.com/cortexlabs/cortex/pkg/lib/sets/strset"
	s "github.com/cortexlabs/cortex/pkg/lib/strings"
)

type ErrorKind int

const (
	ErrUnknown ErrorKind = iota
	ErrUnknownResourceKind
	ErrResourceNotFound
	ErrResourceNameNotFound
	ErrDuplicateResourceName
	ErrDuplicateConfig
	ErrMalformedConfig
	ErrParseConfig
	ErrReadConfig
	ErrMissingDeploymentDefinition
	ErrUndefinedConfig
	ErrUndefinedResource
	ErrResourceWrongType
	ErrSpecifyAllOrNone
	ErrSpecifyOnlyOne
	ErrSpecifyOnlyOneMissing
	ErrOneOfPrerequisitesNotDefined
	ErrCannotBeNull
	ErrK8sQuantityMustBeInt
	ErrBeMoreSpecific
)

var errorKinds = []string{
	"err_unknown",
	"err_unknown_resource_kind",
	"err_resource_not_found",
	"err_resource_name_not_found",
	"err_duplicate_resource_name",
	"err_duplicate_config",
	"err_malformed_config",
	"err_parse_config",
	"err_read_config",
	"err_missing_deployment_definition",
	"err_undefined_config",
	"err_undefined_resource",
	"err_resource_wrong_type",
	"err_specify_all_or_none",
	"err_specify_only_one",
	"err_specify_only_one_missing",
	"err_one_of_prerequisites_not_defined",
	"err_cannot_be_null",
	"err_k8s_quantity_must_be_int",
	"err_be_more_specific",
}

var _ = [1]int{}[int(ErrBeMoreSpecific)-(len(errorKinds)-1)] // Ensure list length matches

func (t ErrorKind) String() string {
	return errorKinds[t]
}

// MarshalText satisfies TextMarshaler
func (t ErrorKind) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *ErrorKind) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(errorKinds); i++ {
		if enum == errorKinds[i] {
			*t = ErrorKind(i)
			return nil
		}
	}

	*t = ErrUnknown
	return nil
}

// UnmarshalBinary satisfies BinaryUnmarshaler
func (t *ErrorKind) UnmarshalBinary(data []byte) error {
	return t.UnmarshalText(data)
}

// MarshalBinary satisfies BinaryMarshaler
func (t ErrorKind) MarshalBinary() ([]byte, error) {
	return []byte(t.String()), nil
}

type Error struct {
	Kind    ErrorKind
	message string
}

func (e Error) Error() string {
	return e.message
}

func ErrorUnknownResourceKind(kind string) error {
	return Error{
		Kind:    ErrUnknownResourceKind,
		message: fmt.Sprintf("invalid resource type %s", s.UserStr(kind)),
	}
}

func ErrorResourceNotFound(name string, resourceType ResourceType) error {
	return Error{
		Kind:    ErrResourceNotFound,
		message: fmt.Sprintf("%s %s not found", resourceType, s.UserStr(name)),
	}
}

func ErrorResourceNameNotFound(name string) error {
	return Error{
		Kind:    ErrResourceNameNotFound,
		message: fmt.Sprintf("resource %s not found", s.UserStr(name)),
	}
}

func ErrorDuplicateResourceName(resources ...Resource) error {
	filePaths := strset.New()
	resourceTypes := strset.New()

	for _, res := range resources {
		resourceTypes.Add(res.GetResourceType().Plural())
		filePaths.Add(res.GetFilePath())
	}

	var pathStrs []string

	if len(filePaths) > 0 {
		pathStrs = append(pathStrs, "defined in "+s.StrsAnd(filePaths.Slice()))
	}

	pathStr := strings.Join(pathStrs, ", ")

	return Error{
		Kind:    ErrDuplicateResourceName,
		message: fmt.Sprintf("name %s must be unique across %s (%s)", s.UserStr(resources[0].GetName()), s.StrsAnd(resourceTypes.Slice()), pathStr),
	}
}

func ErrorDuplicateConfig(resourceType ResourceType) error {
	return Error{
		Kind:    ErrDuplicateConfig,
		message: fmt.Sprintf("%s resource may only be defined once", resourceType.String()),
	}
}

func ErrorMalformedConfig() error {
	return Error{
		Kind:    ErrMalformedConfig,
		message: fmt.Sprintf("cortex YAML configuration files must contain a list of maps"),
	}
}

func ErrorParseConfig() error {
	parseErr := Error{
		Kind:    ErrParseConfig,
		message: fmt.Sprintf("failed to parse config file"),
	}

	return parseErr
}

func ErrorReadConfig() error {
	readErr := Error{
		Kind:    ErrReadConfig,
		message: fmt.Sprintf("failed to read config file"),
	}

	return readErr
}

func ErrorMissingDeploymentDefinition() error {
	return Error{
		Kind:    ErrMissingDeploymentDefinition,
		message: fmt.Sprintf("cortex.yaml must define a %s resource", DeploymentResourceType.String()),
	}
}

func ErrorUndefinedConfig(resourceType ResourceType) error {
	return Error{
		Kind:    ErrUndefinedConfig,
		message: fmt.Sprintf("%s resource is not defined", resourceType.String()),
	}
}

func ErrorUndefinedResource(resourceName string, resourceTypes ...ResourceType) error {
	message := fmt.Sprintf("%s is not defined", s.UserStr(resourceName))

	if len(resourceTypes) == 1 {
		message = fmt.Sprintf("%s %s is not defined", resourceTypes[0].String(), s.UserStr(resourceName))
	} else if len(resourceTypes) > 1 {
		message = fmt.Sprintf("%s is not defined as a %s", s.UserStr(resourceName), s.StrsOr(ResourceTypes(resourceTypes).StringList()))
	}

	if strings.HasPrefix(resourceName, "cortex.") {
		if len(resourceTypes) == 0 {
			message = fmt.Sprintf("%s is not defined in the Cortex namespace", s.UserStr(resourceName))
		} else {
			message = fmt.Sprintf("%s is not defined as a built-in %s in the Cortex namespace", s.UserStr(resourceName), s.StrsOr(ResourceTypes(resourceTypes).StringList()))
		}
	}

	return Error{
		Kind:    ErrUndefinedResource,
		message: message,
	}
}

func ErrorResourceWrongType(resources []Resource, validResourceTypes ...ResourceType) error {
	name := resources[0].GetName()
	resourceTypeStrs := make([]string, len(resources))
	for i, res := range resources {
		resourceTypeStrs[i] = res.GetResourceType().String()
	}

	return Error{
		Kind:    ErrResourceWrongType,
		message: fmt.Sprintf("%s is a %s, but only %s are allowed in this context", s.UserStr(name), s.StrsAnd(resourceTypeStrs), s.StrsOr(ResourceTypes(validResourceTypes).PluralList())),
	}
}

func ErrorSpecifyAllOrNone(vals ...string) error {
	message := fmt.Sprintf("please specify all or none of %s", s.UserStrsAnd(vals))
	if len(vals) == 2 {
		message = fmt.Sprintf("please specify both %s and %s or neither of them", s.UserStr(vals[0]), s.UserStr(vals[1]))
	}

	return Error{
		Kind:    ErrSpecifyAllOrNone,
		message: message,
	}
}

func ErrorSpecifyOnlyOne(vals ...string) error {
	message := fmt.Sprintf("please specify exactly one of %s", s.UserStrsOr(vals))
	if len(vals) == 2 {
		message = fmt.Sprintf("please specify either %s or %s, but not both", s.UserStr(vals[0]), s.UserStr(vals[1]))
	}

	return Error{
		Kind:    ErrSpecifyOnlyOne,
		message: message,
	}
}

func ErrorSpecifyOnlyOneMissing(vals ...string) error {
	message := fmt.Sprintf("please specify one of %s", s.UserStrsOr(vals))
	if len(vals) == 2 {
		message = fmt.Sprintf("please specify either %s or %s", s.UserStr(vals[0]), s.UserStr(vals[1]))
	}

	return Error{
		Kind:    ErrSpecifyOnlyOneMissing,
		message: message,
	}
}

func ErrorOneOfPrerequisitesNotDefined(argName string, prerequisites ...string) error {
	message := fmt.Sprintf("%s specified without specifying %s", s.UserStr(argName), s.UserStrsOr(prerequisites))

	return Error{
		Kind:    ErrOneOfPrerequisitesNotDefined,
		message: message,
	}
}

func ErrorCannotBeNull() error {
	return Error{
		Kind:    ErrCannotBeNull,
		message: "cannot be null",
	}
}

func ErrorK8sQuantityMustBeInt(quantityStr string) error {
	return Error{
		Kind:    ErrK8sQuantityMustBeInt,
		message: fmt.Sprintf("resource compute quantity must be an integer-valued string, e.g. \"2\") (got %s)", s.UserStr(quantityStr)),
	}
}

func ErrorBeMoreSpecific(vals ...string) error {
	return Error{
		Kind:    ErrBeMoreSpecific,
		message: fmt.Sprintf("please specify %s", s.UserStrsOr(vals)),
	}
}

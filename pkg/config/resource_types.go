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
	"strings"
)

type ResourceType int
type ResourceTypes []ResourceType

const (
	UnknownResourceType ResourceType = iota
	DeploymentResourceType
	APIResourceType
)

var (
	types = []string{
		"unknown",
		"deployment",
		"api",
	}

	typePlurals = []string{
		"unknown",
		"deployments",
		"apis",
	}

	resourceTypeAcronyms = map[string]ResourceType{}

	VisibleTypes = ResourceTypes{
		APIResourceType,
	}
)

func ResourceTypeFromString(s string) ResourceType {
	for i := 0; i < len(types); i++ {
		if s == types[i] {
			return ResourceType(i)
		}

		if s == typePlurals[i] {
			return ResourceType(i)
		}

		if t, ok := resourceTypeAcronyms[s]; ok {
			return t
		}
	}
	return UnknownResourceType
}

func ResourceTypeFromKindString(s string) ResourceType {
	for i := 0; i < len(types); i++ {
		if s == types[i] {
			return ResourceType(i)
		}
	}
	return UnknownResourceType
}

func (ts ResourceTypes) String() string {
	return strings.Join(ts.StringList(), ", ")
}

func (ts ResourceTypes) Plural() string {
	return strings.Join(ts.PluralList(), ", ")
}

func (ts ResourceTypes) StringList() []string {
	strs := make([]string, len(ts))
	for i, t := range ts {
		strs[i] = t.String()
	}
	return strs
}

func (ts ResourceTypes) PluralList() []string {
	strs := make([]string, len(ts))
	for i, t := range ts {
		strs[i] = t.Plural()
	}
	return strs
}

func (t ResourceType) String() string {
	return types[t]
}

func (t ResourceType) Plural() string {
	return typePlurals[t]
}

// MarshalText satisfies TextMarshaler
func (t ResourceType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *ResourceType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(types); i++ {
		if enum == types[i] {
			*t = ResourceType(i)
			return nil
		}
	}

	*t = UnknownResourceType
	return nil
}

// UnmarshalBinary satisfies BinaryUnmarshaler
// Needed for msgpack
func (t *ResourceType) UnmarshalBinary(data []byte) error {
	return t.UnmarshalText(data)
}

// MarshalBinary satisfies BinaryMarshaler
func (t ResourceType) MarshalBinary() ([]byte, error) {
	return []byte(t.String()), nil
}

func VisibleResourceTypeFromPrefix(prefix string) (ResourceType, error) {
	prefix = strings.ToLower(prefix)

	if resourceType := ResourceTypeFromString(prefix); resourceType != UnknownResourceType {
		return resourceType, nil
	}

	resourceTypesMap := make(map[ResourceType]struct{})
	for _, resourceType := range VisibleTypes {
		if strings.HasPrefix(resourceType.String(), prefix) {
			resourceTypesMap[resourceType] = struct{}{}
		}

		if strings.HasPrefix(resourceType.Plural(), prefix) {
			resourceTypesMap[resourceType] = struct{}{}
		}
	}

	i := 0
	resourceTypes := make(ResourceTypes, len(resourceTypesMap))
	for resourceType := range resourceTypesMap {
		resourceTypes[i] = resourceType
		i++
	}

	if len(resourceTypes) > 1 {
		return UnknownResourceType, ErrorBeMoreSpecific(resourceTypes.PluralList()...)
	}

	if len(resourceTypes) == 0 {
		return UnknownResourceType, ErrorUnknownResourceKind(prefix)
	}

	return resourceTypes[0], nil
}

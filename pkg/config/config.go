package config

import (
	"fmt"
	"io/ioutil"

	"github.com/cortexlabs/cortex/pkg/lib/cast"
	"github.com/cortexlabs/cortex/pkg/lib/configreader"
	"github.com/cortexlabs/cortex/pkg/lib/errors"
	"github.com/cortexlabs/cortex/pkg/lib/files"
	s "github.com/cortexlabs/cortex/pkg/lib/strings"
)

type Config struct {
	Deployment *Deployment `json:"deployment" yaml:"deployment"`
	APIs       APIs        `json:"apis" yaml:"apis"`
}

func mergeConfigs(target *Config, source *Config) error {
	if source.Deployment != nil {
		if target.Deployment != nil {
			return ErrorDuplicateConfig(DeploymentResourceType)
		}
		target.Deployment = source.Deployment
	}

	target.APIs = append(target.APIs, source.APIs...)

	return nil
}

// var typeFieldValidation = &configreader.StructFieldValidation{
// 	Key: "kind",
// 	Nil: true,
// }

func (config *Config) ValidatePartial() error {
	if config.Deployment != nil {
		if err := config.Deployment.Validate(); err != nil {
			return err
		}
	}

	if config.APIs != nil {
		if err := config.APIs.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (config *Config) Validate() error {
	err := config.ValidatePartial()
	if err != nil {
		return err
	}

	if config.Deployment == nil {
		return ErrorUndefinedConfig(DeploymentResourceType)
	}

	// Check for duplicate names across types that must have unique names
	var resources []Resource
	for _, res := range config.APIs {
		resources = append(resources, res)
	}
	dups := FindDuplicateResourceName(resources...)
	if len(dups) > 0 {
		return ErrorDuplicateResourceName(dups...)
	}

	return nil
}

func (config *Config) MergeBytes(configBytes []byte, filePath string) (*Config, error) {
	sliceData, err := configreader.ReadYAMLBytes(configBytes)
	if err != nil {
		return nil, errors.Wrap(err, filePath)
	}

	subConfig, err := newPartial(sliceData, filePath)
	if err != nil {
		return nil, err
	}

	err = mergeConfigs(config, subConfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func newPartial(configData interface{}, filePath string) (*Config, error) {
	configDataSlice, ok := cast.InterfaceToStrInterfaceMapSlice(configData)
	if !ok {
		return nil, errors.Wrap(ErrorMalformedConfig(), filePath)
	}

	config := &Config{}
	for i, data := range configDataSlice {
		kindInterface, ok := data[KindKey]
		if !ok {
			return nil, errors.Wrap(configreader.ErrorMustBeDefined(), identify(filePath, UnknownResourceType, "", i), KindKey)
		}
		kindStr, ok := kindInterface.(string)
		if !ok {
			return nil, errors.Wrap(configreader.ErrorInvalidPrimitiveType(kindInterface, configreader.PrimTypeString), identify(filePath, UnknownResourceType, "", i), KindKey)
		}

		var errs []error
		resourceType := ResourceTypeFromKindString(kindStr)
		var newResource Resource
		switch resourceType {
		case DeploymentResourceType:
			deployment := &Deployment{}
			errs = configreader.Struct(deployment, data, deploymentValidation)
			config.Deployment = deployment
		case APIResourceType:
			newResource = &API{}
			errs = configreader.Struct(newResource, data, apiValidation)
			if !errors.HasErrors(errs) {
				config.APIs = append(config.APIs, newResource.(*API))
			}
		default:
			return nil, errors.Wrap(ErrorUnknownResourceKind(kindStr), identify(filePath, UnknownResourceType, "", i))
		}

		if errors.HasErrors(errs) {
			name, _ := data[NameKey].(string)
			return nil, errors.Wrap(errors.FirstError(errs...), identify(filePath, resourceType, name, i))
		}

		if newResource != nil {
			newResource.SetIndex(i)
			newResource.SetFilePath(filePath)
		}
	}

	err := config.ValidatePartial()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewPartialPath(filePath string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, filePath, ErrorReadConfig().Error())
	}

	configData, err := configreader.ReadYAMLBytes(configBytes)
	if err != nil {
		return nil, errors.Wrap(err, filePath, ErrorParseConfig().Error())
	}
	return newPartial(configData, filePath)
}

func NewFromBytes(configs map[string][]byte) (*Config, error) {
	var err error
	config := &Config{}
	for filePath, configBytes := range configs {
		if !files.IsFilePathYAML(filePath) {
			continue
		}
		config, err = config.MergeBytes(configBytes, filePath)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func NewFromFiles(filePaths []string) (*Config, error) {
	config := &Config{}
	for _, filePath := range filePaths {
		if !files.IsFilePathYAML(filePath) {
			continue
		}

		configBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, errors.Wrap(err, filePath, ErrorReadConfig().Error())
		}

		config, err = config.MergeBytes(configBytes, filePath)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func ReadDeploymentName(filePath string, relativePath string) (string, error) {
	configBytes, err := files.ReadFileBytes(filePath)
	if err != nil {
		return "", errors.Wrap(err, ErrorReadConfig().Error(), relativePath)
	}
	configData, err := configreader.ReadYAMLBytes(configBytes)
	if err != nil {
		return "", errors.Wrap(err, ErrorParseConfig().Error(), relativePath)
	}
	configDataSlice, ok := cast.InterfaceToStrInterfaceMapSlice(configData)
	if !ok {
		return "", errors.Wrap(ErrorMalformedConfig(), relativePath)
	}

	if len(configDataSlice) == 0 {
		return "", errors.Wrap(ErrorMissingDeploymentDefinition(), relativePath)
	}

	var deploymentName string
	for i, configItem := range configDataSlice {
		kindStr, _ := configItem[KindKey].(string)
		if ResourceTypeFromKindString(kindStr) == DeploymentResourceType {
			if deploymentName != "" {
				return "", errors.Wrap(ErrorDuplicateConfig(DeploymentResourceType), relativePath)
			}

			wrapStr := fmt.Sprintf("%s at %s", DeploymentResourceType.String(), s.Index(i))

			deploymentNameInter, ok := configItem[NameKey]
			if !ok {
				return "", errors.Wrap(configreader.ErrorMustBeDefined(), relativePath, wrapStr, NameKey)
			}

			deploymentName, ok = deploymentNameInter.(string)
			if !ok {
				return "", errors.Wrap(configreader.ErrorInvalidPrimitiveType(deploymentNameInter, configreader.PrimTypeString), relativePath, wrapStr)
			}
			if deploymentName == "" {
				return "", errors.Wrap(configreader.ErrorCannotBeEmpty(), relativePath, wrapStr)
			}
		}
	}

	if deploymentName == "" {
		return "", errors.Wrap(ErrorMissingDeploymentDefinition(), relativePath)
	}

	return deploymentName, nil
}

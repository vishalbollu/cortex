package config

import (
	"github.com/cortexlabs/cortex/pkg/lib/configreader"
	"github.com/cortexlabs/cortex/pkg/providers/k8s"
	"github.com/cortexlabs/cortex/pkg/providers/local"
)

type Provider interface {
	Test() string
	Validate() error
	// CheckLocalDeps() error
	// ValidateObjectPath() error
	// Deploy(config.Config) error
	// Status(string) (string, error)
	// Logs(config.Config, string)
	// Delete(string) error
}

var ProviderStructValidation = &configreader.InterfaceStructValidation{
	TypeKey:         "name",
	TypeStructField: "Name",
	ParsedInterfaceStructTypes: map[interface{}]*configreader.InterfaceStructType{
		LocalProviderType: {
			Type:                   (*local.Provider)(nil),
			StructFieldValidations: local.StructFieldValidations,
		},
		K8SProviderType: {
			Type:                   (*k8s.Provider)(nil),
			StructFieldValidations: k8s.StructFieldValidations,
		},
	},
	Parser: func(str string) (interface{}, error) {
		return ProviderTypeFromString(str), nil
	},
}

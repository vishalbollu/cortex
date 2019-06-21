package config

import (
	"github.com/cortexlabs/cortex/pkg/lib/configreader"
	"github.com/cortexlabs/cortex/pkg/lib/errors"
)

type Deployment struct {
	ResourceFields
	Provider Provider `json:"provider" yaml:"provider"`
}

var deploymentValidation = &configreader.StructValidation{
	StructFieldValidations: []*configreader.StructFieldValidation{
		{
			StructField: "Name",
			StringValidation: &configreader.StringValidation{
				Required:                   true,
				AlphaNumericDashUnderscore: true,
			},
		},
		{
			StructField: "Provider",
			// Key:                       "provider",
			InterfaceStructValidation: ProviderStructValidation,
		},
		// typeFieldValidation,
	},
}

func (api *Deployment) GetResourceType() ResourceType {
	return DeploymentResourceType
}

func (deployment *Deployment) Validate() error {
	if err := deployment.Provider.Validate(); err != nil {
		return errors.Wrap(err, Identify(deployment))
	}
	return nil
}

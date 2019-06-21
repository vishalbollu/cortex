package local

import (
	"github.com/cortexlabs/cortex/pkg/config"
	"github.com/cortexlabs/cortex/pkg/lib/configreader"
)

type Provider struct {
	Name config.ProviderType `json:"name" yaml:"name"`
}

var StructFieldValidations []*configreader.StructFieldValidation = []*configreader.StructFieldValidation{}

func (*Provider) Test() string {
	return "local"
}

func (*Provider) Validate() error {
	return nil
}

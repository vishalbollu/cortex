package k8s

import (
	"github.com/cortexlabs/cortex/pkg/lib/configreader"
	"github.com/cortexlabs/cortex/pkg/lib/k8s"
)

var k8sClient *k8s.Client

func Init() error {
	var err error
	if k8sClient, err = k8s.New("default", false); err != nil {
		return err
	}
	return nil
}

type Provider struct {
	Name      string  `json:"name" yaml:"name"`
	Namespace *string `json:"namespace" yaml:"namespace"`
}

var StructFieldValidations []*configreader.StructFieldValidation = []*configreader.StructFieldValidation{
	{
		StructField: "Namespace",
		StringValidation: &configreader.StringValidation{
			Required:   false,
			Default:    "default",
			AllowEmpty: false,
			DNS1123:    true,
		},
	},
}

func (*Provider) Test() string {
	return "k8s"
}

func (*Provider) Validate() error {
	return nil
}

// var StructSchema *configreader.InterfaceStructType = &configreader.InterfaceStructType{
// 	Type: (*Provider)(nil),
// 	StructFieldValidations: []*configreader.StructFieldValidation{
// 		{
// 			StructField: "Namespace",
// 			StringValidation: &configreader.StringValidation{
// 				Required:   false,
// 				Default:    "default",
// 				AllowEmpty: false,
// 				DNS1123:    true,
// 			},
// 		},
// 	},
// }

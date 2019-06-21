package config

import (
	k8sresource "k8s.io/apimachinery/pkg/api/resource"

	"github.com/cortexlabs/cortex/pkg/lib/configreader"
	"github.com/cortexlabs/cortex/pkg/lib/pointer"
)

type APIs []*API

type API struct {
	ResourceFields
	Model    string    `json:"model" yaml:"model"`
	Replicas int32     `json:"replicas" yaml:"replicas"`
	CPU      *Quantity `json:"cpu" yaml:"cpu"`
	Mem      *Quantity `json:"mem" yaml:"mem"`
	GPU      int64     `json:"gpu" yaml:"gpu"`
}

var apiValidation = &configreader.StructValidation{
	StructFieldValidations: []*configreader.StructFieldValidation{
		{
			StructField: "Name",
			StringValidation: &configreader.StringValidation{
				Required: true,
				DNS1035:  true,
			},
		},
		{
			StructField: "Model",
			StringValidation: &configreader.StringValidation{
				Required:   true,
				AllowEmpty: false,
			},
		},
		{
			StructField: "Replicas",
			Int32Validation: &configreader.Int32Validation{
				Default:     1,
				GreaterThan: pointer.Int32(0),
			},
		},
		{
			StructField: "CPU",
			StringPtrValidation: &configreader.StringPtrValidation{
				Default: nil,
			},
			Parser: QuantityParser(&QuantityValidation{
				Min: k8sresource.MustParse("0"),
			}),
		},
		{
			StructField: "Mem",
			StringPtrValidation: &configreader.StringPtrValidation{
				Default: nil,
			},
			Parser: QuantityParser(&QuantityValidation{
				Min: k8sresource.MustParse("0"),
			}),
		},
		{
			StructField: "GPU",
			Int64Validation: &configreader.Int64Validation{
				Default:              0,
				GreaterThanOrEqualTo: pointer.Int64(0),
			},
		},
		// typeFieldValidation,
	},
}

func (apis APIs) Validate() error {
	for _, api := range apis {
		if err := api.Validate(); err != nil {
			return err
		}
	}

	resources := make([]Resource, len(apis))
	for i, res := range apis {
		resources[i] = res
	}

	dups := FindDuplicateResourceName(resources...)
	if len(dups) > 0 {
		return ErrorDuplicateResourceName(dups...)
	}

	return nil
}

func (api *API) Validate() error {
	return nil
}

func (api *API) GetResourceType() ResourceType {
	return APIResourceType
}

func (apis APIs) Names() []string {
	names := make([]string, len(apis))
	for i, api := range apis {
		names[i] = api.Name
	}
	return names
}

// func (apiCompute *APICompute) ID() string {
// 	var buf bytes.Buffer
// 	buf.WriteString(s.Int32(apiCompute.Replicas))
// 	buf.WriteString(QuantityPtrID(apiCompute.CPU))
// 	buf.WriteString(QuantityPtrID(apiCompute.Mem))
// 	buf.WriteString(s.Int64(apiCompute.GPU))
// 	return hash.Bytes(buf.Bytes())
// }

// func (apiCompute *APICompute) IDWithoutReplicas() string {
// 	var buf bytes.Buffer
// 	buf.WriteString(QuantityPtrID(apiCompute.CPU))
// 	buf.WriteString(QuantityPtrID(apiCompute.Mem))
// 	buf.WriteString(s.Int64(apiCompute.GPU))
// 	return hash.Bytes(buf.Bytes())
// }

// func (apiCompute *APICompute) Equal(apiCompute2 APICompute) bool {
// 	if apiCompute.Replicas != apiCompute2.Replicas {
// 		return false
// 	}
// 	if !QuantityPtrsEqual(apiCompute.CPU, apiCompute2.CPU) {
// 		return false
// 	}
// 	if !QuantityPtrsEqual(apiCompute.Mem, apiCompute2.Mem) {
// 		return false
// 	}

// 	if apiCompute.GPU != apiCompute2.GPU {
// 		return false
// 	}

// 	return true
// }

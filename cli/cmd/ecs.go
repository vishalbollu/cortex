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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/cortexlabs/cortex/pkg/lib/debug"
	"github.com/spf13/cobra"
)

func init() {
	ecsCmd.PersistentFlags()
	addEnvFlag(ecsCmd)
}

// https://medium.com/prodopsio/deploying-fargate-services-using-cloudformation-the-guide-i-wish-i-had-d89b6dc62303

func createTaskDefinition() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))
	svc := ecs.New(sess)
	input := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name:      aws.String("serving-tf"),
				Cpu:       aws.Int64(1024),
				Essential: aws.Bool(true),
				Image:     aws.String("969758392368.dkr.ecr.us-west-2.amazonaws.com/cortexlabs/serving-tf"),
				Memory:    aws.Int64(1024),
				Environment: []*ecs.KeyValuePair{
					&ecs.KeyValuePair{
						Name:  aws.String("AWS_ACCESS_KEY_ID"),
						Value: aws.String("KEY"),
					},
					&ecs.KeyValuePair{
						Name:  aws.String("AWS_SECRET_ACCESS_KEY"),
						Value: aws.String("SECRET"),
					},
					&ecs.KeyValuePair{
						Name:  aws.String("EXTERNAL_MODEL_PATH"),
						Value: aws.String("s3://cortex-examples/iris-model.zip"),
					},
				},
				PortMappings: []*ecs.PortMapping{
					&ecs.PortMapping{
						ContainerPort: aws.Int64(8888),
						HostPort:      aws.Int64(8888),
					},
				},
			},
		},
		Family:           aws.String("iris-api"),
		ExecutionRoleArn: aws.String("arn:aws:iam::969758392368:role/ecsTaskExecutionRole"),
		Cpu:              aws.String("1024"),
		Memory:           aws.String("2048"),
		NetworkMode:      aws.String("awsvpc"),
		RequiresCompatibilities: []*string{
			aws.String("EC2"),
			aws.String("FARGATE"),
		},
	}

	result, err := svc.RegisterTaskDefinition(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				fmt.Println(ecs.ErrCodeServerException, aerr.Error())
			case ecs.ErrCodeClientException:
				fmt.Println(ecs.ErrCodeClientException, aerr.Error())
			case ecs.ErrCodeInvalidParameterException:
				fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}

func createTargetGroup(clusterName string) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
    }))
    
    vpcID := "vpc-0deffdd5a6cd1e8c5"

	svc := ecs.New(sess)
	result, err := svc.ListAttributes(&ecs.ListAttributesInput{
        AttributeName: aws.String("ecs.os-type"),
        AttributeValue:
		Cluster:       aws.String("test-fargate"),
		TargetType:    aws.String("container-instance"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				fmt.Println(ecs.ErrCodeServerException, aerr.Error())
			case ecs.ErrCodeClientException:
				fmt.Println(ecs.ErrCodeClientException, aerr.Error())
			case ecs.ErrCodeInvalidParameterException:
				fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
	debug.Pp("hello")
}

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "ecs an application",
	Long:  "ecs an application.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		// createTaskDefinition()
		createTargetGroup(clusterName)
	},
}

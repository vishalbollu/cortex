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
	"net/url"
	"path"
	"strings"

	"github.com/cortexlabs/cortex/pkg/lib/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/cortexlabs/cortex/pkg/consts"
	"github.com/cortexlabs/cortex/pkg/lib/debug"
	"github.com/cortexlabs/cortex/pkg/lib/errors"

	"github.com/spf13/cobra"
)

func init() {
	serveCmd.PersistentFlags()
	addEnvFlag(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve an application",
	Long:  "serve an application.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func createNamespace(client *kubernetes.Clientset) {
	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "cortex"}}
	_, err := client.CoreV1().Namespaces().Create(nsSpec)
	if err != nil {
		errors.Exit(err)
	}
}

const API_PORT = "8888"
const TF_SERVE_PORT = "9000"
const API_NAME = "iris"

func createDeployment(client *kubernetes.Clientset) {
	dep := k8s.Deployment(&k8s.DeploymentSpec{
		Name:     API_NAME,
		Replicas: 1,
		Labels: map[string]string{
			"apiName": API_NAME,
		},
		Selector: map[string]string{
			"apiName": API_NAME,
		},
		PodSpec: k8s.PodSpec{
			Labels: map[string]string{
				"apiName": API_NAME,
			},
			K8sPodSpec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "api",
						Image:           "<image>",
						ImagePullPolicy: "Always",
						Args: []string{
							"--port=" + API_PORT,
							"--tf-serve-port=" + TF_SERVE_PORT,
							"--model-path=" + "s3://cortex-examples/iris-model.zip",
							"--model-dir=" + path.Join(consts.EmptyDirMountPath, "model"),
						},
						Env:          k8s.AWSCredentials(),
						VolumeMounts: k8s.DefaultVolumeMounts(),
					},
					{
						Name:            "serve",
						Image:           "<image>",
						ImagePullPolicy: "Always",
						Args: []string{
							"--port=" + TF_SERVE_PORT,
							"--model_base_path=" + path.Join(consts.EmptyDirMountPath, "model"),
						},
						Env:          k8s.AWSCredentials(),
						VolumeMounts: k8s.DefaultVolumeMounts(),
					},
				},
				Volumes:            k8s.DefaultVolumes(),
				ServiceAccountName: "default",
			},
		},
		Namespace: NAMESPACE,
	})
	_, err := client.AppsV1beta1().Deployments(NAMESPACE).Create(dep)
	if err != nil {
		errors.Exit(err)
	}

}

func createService(client *kubernetes.Clientset) {
	// svc := k8s.Service(&k8s.ServiceSpec{
	// 	Name:       API_NAME,
	// 	Port:       8888,
	// 	TargetPort: 8888,
	// 	Namespace:  NAMESPACE,
	// })
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      API_NAME,
			Namespace: NAMESPACE,
			Labels: map[string]string{
				"apiName": API_NAME,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"apiName": API_NAME,
			},
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Name:     "http",
					Port:     8888,
					TargetPort: intstr.IntOrString{
						IntVal: 8888,
					},
				},
			},
		},
	}
	_, err := client.CoreV1().Services(NAMESPACE).Create(svc)
	if err != nil {
		errors.Exit(err)
	}
}

const NAMESPACE = "cortex"

func createSecrets(client *kubernetes.Clientset) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "aws-credentials"},
		StringData: map[string]string{
			"AWS_ACCESS_KEY_ID":     "<key>",
			"AWS_SECRET_ACCESS_KEY": "<secret>",
		},
	}

	_, err := client.CoreV1().Secrets(NAMESPACE).Create(secret)
	if err != nil {
		errors.Exit(err)
	}
}

func serve() {
	kubeConfig := path.Join(homedir.HomeDir(), ".kube", "config")
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		errors.Exit(err)
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		errors.Exit(err)
	}

	// createNamespace(client)
	// createSecrets(client)
	// createDeployment(client)
	// createService(client)

	hostURL, _ := url.Parse(restConfig.Host)
	service, err := client.CoreV1().Services(NAMESPACE).Get(API_NAME, metav1.GetOptions{})
	// if k8serrors.IsNotFound(err) {
	// 	return nil, nil
	// }
	// if err != nil {
	// 	return nil, errors.WithStack(err)
	// }
	// service.TypeMeta = serviceTypeMeta
	for _, port := range service.Spec.Ports {
		if port.Name == "http" {
			debug.Pp(fmt.Sprintf("http://%s:%d", strings.Split(hostURL.Host, ":")[0], port.NodePort))
		}
	}
}

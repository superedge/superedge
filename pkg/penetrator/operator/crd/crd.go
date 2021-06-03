/*
Copyright 2020 The SuperEdge Authors.
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

package crd

import (
	"github.com/superedge/superedge/pkg/penetrator/apis/nodetask.apps.superedge.io/v1beta1"
	"github.com/superedge/superedge/pkg/penetrator/operator/context"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	NodesTaskCustomResourceDefinition = apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nodetasks.nodetask.apps.superedge.io",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: v1beta1.SchemeGroupVersion.Group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:       "NodeTask",
				Plural:     "nodetasks",
				ShortNames: []string{"nt"},
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1beta1",
					Storage: true,
					Served:  true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"nodeNamePrefix": {
											Type:        "string",
											Description: "Node name prefix",
										},
										"targetMachines": {
											Type: "array",
											Items: &apiextensionsv1.JSONSchemaPropsOrArray{
												Schema: &apiextensionsv1.JSONSchemaProps{
													Type:   "string",
													Format: "ipv4",
												},
											},
											Description: "Install the ip list of the node",
										},
										"nodeNamesOverride": {
											Type: "object",
											AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
												Schema: &apiextensionsv1.JSONSchemaProps{
													Type: "string",
												},
											},
											Description: "Specify the node name and ip mapping",
										},
										"proxyNode": {
											Type:        "string",
											Description: "The name of the node in the cluster running the add node job",
										},
										"sshCredential": {
											Type:        "string",
											Description: "The name of the secret that stores the ssh password or private key",
										},
										"sshPort": {
											Type:        "integer",
											Description: "SSH login port",
											Default:     &apiextensionsv1.JSON{[]byte("22")},
										},
									},
									Required: []string{
										"sshCredential",
										"proxyNode",
									},
								},
								"status": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"nodeStatus": {
											Type: "object",
											AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
												Schema: &apiextensionsv1.JSONSchemaProps{
													Type: "string",
												},
											},
											Description: "Nodes that have not been installed",
										},
										"nodetaskStatus": {
											Type: "string",
											Enum: []apiextensionsv1.JSON{
												{Raw: []byte(`"creating"`)},
												{Raw: []byte(`"ready"`)},
											},
											Description: "The execution status of nodetask",
										},
									},
								},
							},
						},
					},
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
				},
			},
			Scope: apiextensionsv1.ClusterScoped,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{},
		},
	}
)

//Create crd when operator starts
func InstallWithMaxRetry(ctx *context.NodeTaskContext, client clientset.Interface, crd *apiextensionsv1.CustomResourceDefinition, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		if err = Install(ctx, client, crd); err == nil {
			return nil
		}
	}
	return err
}

func Install(ctx *context.NodeTaskContext, client clientset.Interface, crd *apiextensionsv1.CustomResourceDefinition) error {
	existing, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		_, err = client.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		crd.ResourceVersion = existing.ResourceVersion
		_, err = client.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

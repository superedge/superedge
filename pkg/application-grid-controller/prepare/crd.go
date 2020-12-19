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

package prepare

import (
	"context"
	"fmt"
	"superedge/pkg/application-grid-controller/controller/common"
	"superedge/pkg/util/kubeclient"
	"gopkg.in/yaml.v2"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"time"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

type crdPreparator struct {
	client clientset.Interface
}

func NewCRDPreparator(client clientset.Interface) *crdPreparator {
	return &crdPreparator{client: client}
}

func (p *crdPreparator) Prepare(crds ...schema.GroupVersionKind) error {
	if len(crds) == 0 {
		return nil
	}

	ready, err := p.getReadyCRDs()
	if err != nil {
		return err
	}

	crdStatus := map[schema.GroupVersionKind]*apiext.CustomResourceDefinition{}
	for _, crdDef := range crds {
		/*create crd
		 */
		crd, err := p.createCRD(crdDef, ready)
		if err != nil {
			return err
		}
		crdStatus[crdDef] = crd
	}

	ready, err = p.getReadyCRDs()
	if err != nil {
		return err
	}

	for gvk, crd := range crdStatus {
		if readyCrd, ok := ready[crd.Name]; ok {
			crdStatus[gvk] = readyCrd
		} else {
			if err := p.waitCRD(crd.Name, gvk, crdStatus); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *crdPreparator) createCRD(gvk schema.GroupVersionKind, ready map[string]*apiext.CustomResourceDefinition) (*apiext.CustomResourceDefinition, error) {
	var (
		crd  *apiext.CustomResourceDefinition
		data []byte
		err  error
	)
	//crd := util.ToCustomResourceDefinition(gvk)

	if gvk.Kind == "DeploymentGrid" {
		data, err = kubeclient.ParseString(common.DeploymentGridCRDYaml, map[string]interface{}{})
		if err != nil {
			return nil, err
		}
	}

	if gvk.Kind == "ServiceGrid" {
		data, err = kubeclient.ParseString(common.ServiceGridCRDYaml, map[string]interface{}{})
		if err != nil {
			return nil, err
		}
	}

	//klog.V(8).Infof("Create yaml: %s", string(data))
	objBytes := data
	obj := new(object)
	err = yaml.Unmarshal(objBytes, obj)
	if err != nil {
		return nil, err
	}
	if obj.Kind != "" {
		crd = new(apiext.CustomResourceDefinition)
		if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), objBytes, crd); err != nil {
			return nil, err
		}
	}

	if crd != nil {
		existing, ok := ready[crd.Name]
		if ok {
			if !equality.Semantic.DeepEqual(crd.Spec.Validation, existing.Spec.Validation) ||
				!equality.Semantic.DeepEqual(crd.Spec.Versions, existing.Spec.Versions) {
				existing.Spec = crd.Spec
				klog.Infof("Updating CRD %s", crd.Name)
				return p.client.ApiextensionsV1beta1().CustomResourceDefinitions().Update(context.TODO(), existing, metav1.UpdateOptions{})
			}
			return existing, nil
		}

		klog.Infof("Creating CRD %s", crd.Name)
		if newCrd, err := p.client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{}); errors.IsAlreadyExists(err) {
			return p.client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(context.TODO(), crd.Name, metav1.GetOptions{})
		} else if err != nil {
			return nil, err
		} else {
			return newCrd, nil
		}
	}
	return nil, fmt.Errorf("crd is nil")
}

func (p *crdPreparator) getReadyCRDs() (map[string]*apiext.CustomResourceDefinition, error) {
	list, err := p.client.ApiextensionsV1beta1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ready := map[string]*apiext.CustomResourceDefinition{}
	for i, crd := range list.Items {
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiext.Established:
				if cond.Status == apiext.ConditionTrue {
					ready[crd.Name] = &list.Items[i]
				}
			}
		}
	}
	return ready, nil
}

func (p *crdPreparator) waitCRD(name string, gvk schema.GroupVersionKind,
	crdStatus map[schema.GroupVersionKind]*apiext.CustomResourceDefinition) error {
	klog.Infof("Waiting for CRD %s to become available", name)
	defer klog.Infof("Done waiting for CRD %s to become available", name)

	first := true
	return wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		if !first {
			klog.Infof("Waiting for CRD %s to become available", name)
		}
		first = false

		crd, err := p.client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiext.Established:
				if cond.Status == apiext.ConditionTrue {
					crdStatus[gvk] = crd
					return true, err
				}
			case apiext.NamesAccepted:
				if cond.Status == apiext.ConditionFalse {
					klog.Infof("Name conflict on %s: %v\n", name, cond.Reason)
				}
			}
		}

		return false, nil
	})
}

type object struct {
	Kind string `yaml:"kind"`
}

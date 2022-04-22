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
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	"io/ioutil"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"time"

	"fmt"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	DeploymentGridCRDYaml  = "/etc/superedge/application-grid-controller/manifests/superedge.io_deploymentgrids.yaml"
	StatefulSetGridCRDYaml = "/etc/superedge/application-grid-controller/manifests/superedge.io_statefulsetgrids.yaml"
	ServiceGridCRDYaml     = "/etc/superedge/application-grid-controller/manifests/superedge.io_servicegrids.yaml"
)

type crdPreparator struct {
	client clientset.Interface
}

func NewCRDPreparator(client clientset.Interface) *crdPreparator {
	return &crdPreparator{client: client}
}

func (p *crdPreparator) Prepare(stopCh <-chan struct{}, gvks ...schema.GroupVersionKind) error {
	if len(gvks) == 0 {
		return nil
	}
	// First of all, create or update edge CRDs
	err := p.prepareCRDs(gvks...)
	if err != nil {
		return err
	}
	// Loop background
	go wait.Until(func() {
		p.prepareCRDs(gvks...)
	}, time.Minute, stopCh)
	return nil
}

func (p *crdPreparator) prepareCRDs(gvks ...schema.GroupVersionKind) error {
	// create or update specified edge CRDs
	for _, gvk := range gvks {
		curCRD, err := p.createOrUpdateCRD(gvk)
		if err != nil {
			return err
		}
		if err := p.waitCRD(curCRD.Name); err != nil {
			return err
		}
	}
	return nil
}

// createOrUpdateCRDs create or update specified edge CRDs
func (p *crdPreparator) createOrUpdateCRD(gvk schema.GroupVersionKind) (*apiext.CustomResourceDefinition, error) {
	klog.V(5).Infof("Trying to create or update GroupVersionKind %s CRD", gvk)
	defer klog.V(5).Infof("Done creating or updating GroupVersionKind %s CRD", gvk)
	var (
		crdBytes []byte
		err      error
	)
	// create specified GroupVersionKind edge CRD
	switch gvk.Kind {
	case common.DeploymentGridKind:
		f, err := ioutil.ReadFile(DeploymentGridCRDYaml)
		if err != nil {
			return nil, err
		}
		crdBytes, err = kubeclient.ParseString(string(f), map[string]interface{}{})
	case common.StatefulSetGridKind:
		f, err := ioutil.ReadFile(StatefulSetGridCRDYaml)
		if err != nil {
			return nil, err
		}
		crdBytes, err = kubeclient.ParseString(string(f), map[string]interface{}{})
	case common.ServiceGridKind:
		f, err := ioutil.ReadFile(ServiceGridCRDYaml)
		if err != nil {
			return nil, err
		}
		crdBytes, err = kubeclient.ParseString(string(f), map[string]interface{}{})
	default:
		err = fmt.Errorf("Invalid edge group version kind resource %s", gvk.Kind)
	}
	if err != nil {
		return nil, err
	}

	crd := new(apiext.CustomResourceDefinition)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), crdBytes, crd); err != nil {
		return nil, err
	}
	// create or update relevant edge CRD
	curCRD, err := p.client.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crd.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// try to create edge CRD
		klog.V(4).Infof("Creating CRD %s", crd.Name)
		if newCrd, err := p.client.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{}); errors.IsAlreadyExists(err) {
			return p.client.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crd.Name, metav1.GetOptions{})
		} else if err != nil {
			return nil, err
		} else {
			return newCrd, nil
		}
	}
	// update edge CRD if necessary
	if !equality.Semantic.DeepEqual(crd.Spec.Versions, curCRD.Spec.Versions) ||
		!equality.Semantic.DeepEqual(crd.Spec.Versions, curCRD.Spec.Versions) {
		curCRD.Spec = crd.Spec
		klog.V(4).Infof("Updating CRD %s", crd.Name)
		return p.client.ApiextensionsV1().CustomResourceDefinitions().Update(context.TODO(), curCRD, metav1.UpdateOptions{})
	}
	return curCRD, nil
}

// waitCRD waits for specified edge CRD to become available
func (p *crdPreparator) waitCRD(name string) error {
	klog.V(5).Infof("Waiting for CRD %s to become available", name)
	defer klog.V(5).Infof("Done waiting for CRD %s to become available", name)

	first := true
	return wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		if !first {
			klog.V(5).Infof("Waiting for CRD %s to become available", name)
		}
		first = false

		crd, err := p.client.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiext.Established:
				if cond.Status == apiext.ConditionTrue {
					return true, err
				}
			case apiext.NamesAccepted:
				if cond.Status == apiext.ConditionFalse {
					klog.Infof("Name conflict on %s: %v", name, cond.Reason)
				}
			}
		}

		return false, nil
	})
}

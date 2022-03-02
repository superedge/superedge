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

package deployment

import (
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
)

func (dgc *DeploymentGridController) addNameSpace(obj interface{}) {
	ns := obj.(*corev1.Namespace)
	klog.V(4).Infof("Adding NameSpace %s", ns.Name)

	if _, ok := ns.Labels[common.FedManagedClustIdKey]; !ok {
		return
	}

	labelSelector := labels.NewSelector()
	gridRequirement, err := labels.NewRequirement(common.FedrationKey, selection.Equals, []string{"yes"})
	if err != nil {
		klog.V(4).Infof("Adding NameSpace %s error: gererate requirement err %v", ns.Name, err)
		return
	}
	labelSelector = labelSelector.Add(*gridRequirement)

	dgList, err := dgc.dpGridLister.List(labelSelector)
	if err != nil {
		klog.V(4).Infof("Adding NameSpace %s error: get dgList err %v", ns.Name, err)
		return
	}

	for _, dg := range dgList {
		dgc.enqueueDeploymentGrid(dg)
	}
}

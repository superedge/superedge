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

package service

import (
	"fmt"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
)

func (sgc *ServiceGridController) addServiceGrid(obj interface{}) {
	sg := obj.(*crdv1.ServiceGrid)
	klog.V(4).Infof("Adding service grid %s", sg.Name)
	sgc.enqueueServiceGrid(sg)
}

func (sgc *ServiceGridController) updateServiceGrid(oldObj, newObj interface{}) {
	old := oldObj.(*crdv1.ServiceGrid)
	cur := newObj.(*crdv1.ServiceGrid)
	klog.V(4).Infof("Updating service grid %s", old.Name)
	sgc.enqueueServiceGrid(cur)
}

func (sgc *ServiceGridController) deleteServiceGrid(obj interface{}) {
	sg, ok := obj.(*crdv1.ServiceGrid)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		sg, ok = tombstone.Obj.(*crdv1.ServiceGrid)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a Service %#v", obj))
			return
		}
	}
	klog.V(4).Infof("Deleting service grid %s", sg.Name)
	sgc.enqueueServiceGrid(sg)
}

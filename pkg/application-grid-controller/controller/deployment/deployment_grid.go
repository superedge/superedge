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
	"fmt"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	crdv1 "superedge/pkg/application-grid-controller/apis/superedge.io/v1"
)

func (dgc *DeploymentGridController) addDeploymentGrid(obj interface{}) {
	dg := obj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Adding deployment grid %s", dg.Name)
	dgc.enqueueDeploymentGrid(dg)
}

func (dgc *DeploymentGridController) updateDeploymentGrid(oldObj, newObj interface{}) {
	old := oldObj.(*crdv1.DeploymentGrid)
	cur := newObj.(*crdv1.DeploymentGrid)
	klog.V(4).Infof("Updating deployment grid %s", old.Name)
	dgc.enqueueDeploymentGrid(cur)
}

func (dgc *DeploymentGridController) deleteDeploymentGrid(obj interface{}) {
	dg, ok := obj.(*crdv1.DeploymentGrid)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		dg, ok = tombstone.Obj.(*crdv1.DeploymentGrid)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}
	klog.V(4).Infof("Deleting deployment grid %s", dg.Name)
	dgc.enqueueDeploymentGrid(dg)
}

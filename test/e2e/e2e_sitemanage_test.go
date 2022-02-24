/*
Copyright 2022 The SuperEdge Authors.

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

package e2e

import (
	"context"
	//crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"testing"
)

func TestApplyNodeGroup(t *testing.T) {
	ctx := context.Background()
	clientSet, _, crdclient := initClient(t)

	sharedInformerFactory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := sharedInformerFactory.Core().V1().Nodes()

	nodelist, err := ReadyNodes(ctx, clientSet, nodeInformer, "zone1=nodeunit1")
	t.Log(len(nodelist))
	if err != nil {
		t.Fatal("check node list fail", err)
	}

	nodegroupObj := &v1alpha1.NodeGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nodegroup-e2e",
			Namespace: "default",
		},
		Spec: v1alpha1.NodeGroupSpec{
			AutoFindNodeKeys: []string{"zone1"},
		},
	}
	_, err = crdclient.SiteV1alpha1().NodeGroups().Create(ctx, nodegroupObj, metav1.CreateOptions{})

	if err != nil {
		t.Fatal("create nodegroup fail", err)
	}
	defer crdclient.SiteV1alpha1().NodeGroups().Delete(ctx, "nodegroup-e2e", metav1.DeleteOptions{})

	// check is it have deployment deploymentgrid-e2e-nodeunitname
	result, err := waitForNodeUnit(t, ctx, "nodeunit1", crdclient)
	if !result {
		t.Fatal("nodeunit not found", err)
	}
}

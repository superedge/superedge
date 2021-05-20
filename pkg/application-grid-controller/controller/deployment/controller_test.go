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
	"flag"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/deployment/util"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/testutil"
	crdfake "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned/fake"
	crdinformers "github.com/superedge/superedge/pkg/application-grid-controller/generated/informers/externalversions"
)

func init() {
	flagsets := flag.NewFlagSet("test", flag.ExitOnError)
	klog.InitFlags(flagsets)
	_ = flagsets.Set("v", "4")
	_ = flagsets.Set("logtostderr", "true")
	_ = flagsets.Parse(nil)
}

type fixture struct {
	t testing.TB

	kubeClient *fake.Clientset
	crdClient  *crdfake.Clientset

	// Objects to put in the store.
	dpLister   []*appsv1.Deployment
	dpgLister  []*crdv1.DeploymentGrid
	nodeLister []*corev1.Node

	// Actions expected to happen on the client. Objects from here are also
	// preloaded into NewSimpleFake.
	actions    []core.Action
	objects    []runtime.Object
	crdObjects []runtime.Object
}

var (
	alwaysReady = func() bool { return true }
)

func newFixture(t testing.TB) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.crdObjects = []runtime.Object{}
	return f
}

func (f *fixture) newController() (*DeploymentGridController, informers.SharedInformerFactory, crdinformers.SharedInformerFactory) {
	f.kubeClient = fake.NewSimpleClientset(f.objects...)
	f.crdClient = crdfake.NewSimpleClientset(f.crdObjects...)

	kubeFactory := informers.NewSharedInformerFactory(f.kubeClient, 0)
	crdFactory := crdinformers.NewSharedInformerFactory(f.crdClient, 0)

	dpGridInformer := crdFactory.Superedge().V1().DeploymentGrids()
	dpInformer := kubeFactory.Apps().V1().Deployments()
	nodeInformer := kubeFactory.Core().V1().Nodes()

	c := NewDeploymentGridController(dpGridInformer, dpInformer, nodeInformer, f.kubeClient, f.crdClient)
	c.eventRecorder = &record.FakeRecorder{}
	c.dpListerSynced = alwaysReady
	c.dpGridListerSynced = alwaysReady
	c.nodeListerSynced = alwaysReady

	for _, d := range f.dpLister {
		err := dpInformer.Informer().GetIndexer().Add(d)
		if err != nil {
			f.t.Errorf("Add dp err: %v", err)
		}
	}
	for _, dpg := range f.dpgLister {
		err := dpGridInformer.Informer().GetIndexer().Add(dpg)
		if err != nil {
			f.t.Errorf("Add dpGrid err: %v", err)
		}
	}
	for _, n := range f.nodeLister {
		err := nodeInformer.Informer().GetIndexer().Add(n)
		if err != nil {
			f.t.Errorf("Add node err: %v", err)
		}
	}
	return c, kubeFactory, crdFactory
}

func (f *fixture) runExpectError(deploymentGridName string, startInformers bool) {
	f.run_(deploymentGridName, startInformers, true)
}

func (f *fixture) run(deploymentGridName string) {
	f.run_(deploymentGridName, true, false)
}

func (f *fixture) run_(deploymentGridName string, startInformers bool, expectError bool) {
	c, kubeInformer, crdInformer := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		kubeInformer.Start(stopCh)
		crdInformer.Start(stopCh)
	}
	err := c.syncDeploymentGrid(deploymentGridName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing deployment grid: %v", err)
		return
	}

	if expectError && err == nil {
		f.t.Error("expected error syncing deployment grid, got nil")
	}

	actions := filterInformerActions(append(f.kubeClient.Actions(), f.crdClient.Actions()...))
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) && action.GetSubresource() == expectedAction.GetSubresource()) {
			f.t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}
}

func filterInformerActions(actions []core.Action) []core.Action {
	ret := make([]core.Action, 0)
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "deployments") ||
				action.Matches("list", "deploymentgrids") ||
				action.Matches("list", "nodes") ||
				action.Matches("watch", "deployments") ||
				action.Matches("watch", "deploymentgrids") ||
				action.Matches("watch", "nodes")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func newDeploymentGrid(name string, replicas int, gridUniqKey string, selector map[string]string) *crdv1.DeploymentGrid {
	dpg := &crdv1.DeploymentGrid{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "superedge.io/v1",
			Kind:       "DeploymentGrid",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:         uuid.NewUUID(),
			Name:        name,
			Namespace:   metav1.NamespaceDefault,
			Annotations: make(map[string]string),
		},
		Spec: crdv1.DeploymentGridSpec{
			GridUniqKey: gridUniqKey,
			Template: appsv1.DeploymentSpec{
				Replicas: func() *int32 { i := int32(replicas); return &i }(),
				Selector: &metav1.LabelSelector{MatchLabels: selector},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: selector,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "foo/bar",
							},
						},
					},
				},
			},
		},
	}
	return dpg
}

func newNode(name string, labels map[string]string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
			UID:    uuid.NewUUID(),
		},
	}
}

func newDeployment(dpg *crdv1.DeploymentGrid, name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			UID:       uuid.NewUUID(),
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				common.GridSelectorName: dpg.Name,
			},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(dpg, util.ControllerKind)},
		},
		Spec: dpg.Spec.Template,
	}
}

func (f *fixture) expectGetDeploymentGridAction(o *crdv1.DeploymentGrid) {
	action := core.NewGetAction(schema.GroupVersionResource{Resource: "deploymentgrids"}, o.Namespace, o.Name)
	f.actions = append(f.actions, action)
}

func (f *fixture) expectUpdateDeploymentGridStatusAction(o *crdv1.DeploymentGrid) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Resource: "deploymentgrids"}, o.Namespace, o)
	action.Subresource = "status"
	f.actions = append(f.actions, action)
}

func (f *fixture) expectCreateDPAction(dp *appsv1.Deployment) {
	f.actions = append(f.actions, core.NewCreateAction(schema.GroupVersionResource{Resource: "deployments"}, dp.Namespace, dp))
}

func TestSyncDeploymentGridCreateNoDeployment(t *testing.T) {
	f := newFixture(t)

	n := newNode("nolabel", nil)
	f.nodeLister = append(f.nodeLister, n)
	f.objects = append(f.objects, n)

	dpg := newDeploymentGrid("foo", 1, "zone", map[string]string{"foo": "bar"})
	f.dpgLister = append(f.dpgLister, dpg)
	f.crdObjects = append(f.crdObjects, dpg)

	f.expectUpdateDeploymentGridStatusAction(dpg)
	f.run(testutil.GetKey(dpg, t))
}

func TestSyncDeploymentGridCreateDeployment(t *testing.T) {
	f := newFixture(t)

	nodes := []*corev1.Node{
		newNode("zone-1", map[string]string{"zone": "1"}),
		newNode("zone-2", map[string]string{"zone": "2"}),
		newNode("zone-3", map[string]string{"zone": "3"}),
	}

	for _, n := range nodes {
		f.nodeLister = append(f.nodeLister, n)
		f.objects = append(f.objects, n)
	}

	dpg := newDeploymentGrid("foo", 1, "zone", map[string]string{"foo": "bar"})
	f.dpgLister = append(f.dpgLister, dpg)
	f.crdObjects = append(f.crdObjects, dpg)

	f.expectCreateDPAction(newDeployment(dpg, "foo-1"))
	f.expectCreateDPAction(newDeployment(dpg, "foo-2"))
	f.expectCreateDPAction(newDeployment(dpg, "foo-3"))

	f.expectUpdateDeploymentGridStatusAction(dpg)

	f.run(testutil.GetKey(dpg, t))
}

func TestSyncDeploymentGridDeletionRace(t *testing.T) {
	f := newFixture(t)

	dpg := newDeploymentGrid("foo", 1, "zone", map[string]string{"foo": "bar"})
	dpg2 := *dpg
	// Lister (cache) says NOT deleted.
	f.dpgLister = append(f.dpgLister, dpg)

	// Bare client says it IS deleted. This should be presumed more up-to-date.
	now := metav1.Now()
	dpg2.DeletionTimestamp = &now
	f.crdObjects = append(f.crdObjects, &dpg2)

	// The recheck is only triggered if a matching orphan exists.
	dp := newDeployment(dpg, "zone-1")
	dp.OwnerReferences = nil
	f.objects = append(f.objects, dp)
	f.dpLister = append(f.dpLister, dp)

	// Expect to only recheck DeletionTimestamp.
	f.expectGetDeploymentGridAction(dpg)
	// Sync should fail and requeue to let cache catch up.
	// Don't start informers, since we don't want cache to catch up for this test.
	f.runExpectError(testutil.GetKey(dpg, t), false)
}

func TestDontSyncDeploymentGridWithEmptyGridUniqKey(t *testing.T) {
	f := newFixture(t)

	dpg := newDeploymentGrid("foo", 1, "", map[string]string{"foo": "bar"})
	f.dpgLister = append(f.dpgLister, dpg)
	f.crdObjects = append(f.crdObjects, dpg)

	// Normally there should be a status update but the fake deployment grid
	// has gridUniqKey set so there is no action happening here.
	f.run(testutil.GetKey(dpg, t))
}

func TestDeploymentDeletionEnqueuesRecreateDeployment(t *testing.T) {
	f := newFixture(t)

	nodes := []*corev1.Node{
		newNode("zone-1", map[string]string{"zone": "1"}),
		newNode("zone-2", map[string]string{"zone": "2"}),
		newNode("zone-3", map[string]string{"zone": "3"}),
	}

	for _, n := range nodes {
		f.nodeLister = append(f.nodeLister, n)
		f.objects = append(f.objects, n)
	}

	dpg := newDeploymentGrid("foo", 1, "zone", map[string]string{"foo": "bar"})
	f.dpgLister = append(f.dpgLister, dpg)
	f.crdObjects = append(f.crdObjects, dpg)

	existedDps := []*appsv1.Deployment{
		newDeployment(dpg, "foo-1"),
		newDeployment(dpg, "foo-2"),
		newDeployment(dpg, "foo-3"),
	}
	f.dpLister = append(f.dpLister, existedDps...)

	c, _, _ := f.newController()
	enqueued := false
	c.enqueueDeploymentGrid = func(o *crdv1.DeploymentGrid) {
		if o.Name == "foo" {
			enqueued = true
		}
	}

	c.deleteDeployment(existedDps[0])

	if !enqueued {
		t.Errorf("expected deployment grid %q to be queued after deployment deletion", dpg.Name)
		return
	}
}

func TestGetDeploymentsForDeploymentGrid(t *testing.T) {
	f := newFixture(t)

	nodes := []*corev1.Node{
		newNode("zone-1", map[string]string{"zone": "1"}),
		newNode("zone-2", map[string]string{"zone": "2"}),
		newNode("zone-3", map[string]string{"zone": "3"}),
	}

	for _, n := range nodes {
		f.nodeLister = append(f.nodeLister, n)
		f.objects = append(f.objects, n)
	}

	dpg := newDeploymentGrid("foo", 1, "zone", map[string]string{"foo": "bar"})
	f.dpgLister = append(f.dpgLister, dpg)
	f.crdObjects = append(f.crdObjects, dpg)

	existedDps := []*appsv1.Deployment{
		newDeployment(dpg, "foo-1"),
		newDeployment(dpg, "foo-2"),
		newDeployment(dpg, "foo-3"),
	}
	f.dpLister = append(f.dpLister, existedDps...)

	c, coreInformer, _ := f.newController()
	stopCh := make(chan struct{})
	defer close(stopCh)
	coreInformer.Start(stopCh)

	dpList, err := c.getDeploymentForGrid(dpg)
	if err != nil {
		t.Errorf("getDeploymentForGrid() %v", err)
		return
	}

	nameSets := sets.NewString()
	for _, dp := range dpList {
		nameSets.Insert(dp.Name)
	}

	for _, dp := range existedDps {
		if !nameSets.Has(dp.Name) {
			t.Errorf("can't find %s", dp.Name)
			return
		}
	}
}

func TestUpdateDeploymentChangeControllerRef(t *testing.T) {
	f := newFixture(t)

	nodes := []*corev1.Node{
		newNode("zone-1", map[string]string{"zone": "1"}),
		newNode("zone-2", map[string]string{"zone": "2"}),
		newNode("zone-3", map[string]string{"zone": "3"}),
	}

	for _, n := range nodes {
		f.nodeLister = append(f.nodeLister, n)
		f.objects = append(f.objects, n)
	}

	dpg1 := newDeploymentGrid("foo", 1, "zone", nil)
	dpg2 := newDeploymentGrid("bar", 1, "zone", nil)

	dp := newDeployment(dpg1, "foo-1")

	f.dpgLister = append(f.dpgLister, dpg1, dpg2)
	f.dpLister = append(f.dpLister, dp)
	f.objects = append(f.objects, dp)
	f.crdObjects = append(f.crdObjects, dpg1, dpg2)

	// Create the fixture but don't start it,
	// so nothing happens in the background.
	c, _, _ := f.newController()
	// Change ControllerRef and expect both old and new to queue.
	prev := *dp
	prev.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(dpg2, util.ControllerKind)}
	next := *dp
	c.updateDeployment(&prev, &next)
	if got, want := c.queue.Len(), 2; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
}

func TestUpdateDeploymentOrphanWithNewLabels(t *testing.T) {
	f := newFixture(t)

	nodes := []*corev1.Node{
		newNode("zone-1", map[string]string{"zone": "1"}),
		newNode("zone-2", map[string]string{"zone": "2"}),
		newNode("zone-3", map[string]string{"zone": "3"}),
	}

	for _, n := range nodes {
		f.nodeLister = append(f.nodeLister, n)
		f.objects = append(f.objects, n)
	}

	dpg1 := newDeploymentGrid("foo", 1, "zone", nil)
	dpg2 := newDeploymentGrid("bar", 1, "zone", nil)

	dp := newDeployment(dpg1, "foo-1")
	dp.OwnerReferences = nil

	f.dpgLister = append(f.dpgLister, dpg1, dpg2)
	f.dpLister = append(f.dpLister, dp)
	f.objects = append(f.objects, dp)
	f.crdObjects = append(f.crdObjects, dpg1, dpg2)

	// Create the fixture but don't start it,
	// so nothing happens in the background.
	c, _, _ := f.newController()
	// Change ControllerRef and expect both old and new to queue.
	prev := *dp
	prev.Labels = map[string]string{common.GridSelectorName: "nobar"}
	next := *dp
	c.updateDeployment(&prev, &next)
	if got, want := c.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
}

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

package storage

import (
	"sort"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

//func init() {
//	flagsets := flag.NewFlagSet("test", flag.ExitOnError)
//	klog.InitFlags(flagsets)
//	_ = flagsets.Set("v", "4")
//	_ = flagsets.Set("logtostderr", "true")
//	_ = flagsets.Parse(nil)
//}
//
//func TestCache(t *testing.T) {
//	svcCh := make(chan watch.Event)
//	endpointBroadcaster := watch.NewLongQueueBroadcaster(1000, watch.DropIfChannelFull)
//	endpointSliceChV1 := make(chan watch.Event)
//	endpointSliceChV1Beta1 := make(chan watch.Event)
//	nodeBroadcaster := watch.NewLongQueueBroadcaster(1000, watch.DropIfChannelFull)
//	supportEndpointSlice := true
//	stop := false
//	defer func() {
//		stop = true
//	}()
//
//	epsWatch := endpointBroadcaster.Watch()
//	nodeWatch := nodeBroadcaster.Watch()
//	go func() {
//		for !stop {
//			select {
//			case <-svcCh:
//			case <-epsWatch.ResultChan():
//			case <-nodeWatch.ResultChan():
//
//			}
//		}
//	}()
//
//	cache := NewStorageCache("hostname", true, false, svcCh, endpointSliceChV1, endpointSliceChV1Beta1, endpointBroadcaster, nodeBroadcaster, supportEndpointSlice)
//
//	testNodes := make([]*v1.Node, 10)
//	nodeEventHandler := cache.NodeEventHandler()
//	for i := 0; i < len(testNodes); i++ {
//		testNodes[i] = &v1.Node{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:   fmt.Sprintf("node-%d", i),
//				Labels: make(map[string]string),
//			},
//		}
//		nodeEventHandler.OnAdd(testNodes[i])
//	}
//
//	testServices := make([]*v1.Service, 10)
//	serviceEventHandler := cache.ServiceEventHandler()
//	for i := 0; i < len(testServices); i++ {
//		testServices[i] = &v1.Service{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:        fmt.Sprintf("%d", i),
//				Namespace:   metav1.NamespaceDefault,
//				Annotations: make(map[string]string),
//			},
//		}
//		if i%2 == 0 {
//			testServices[i].Annotations[TopologyAnnotationsKey] = "[\"foo\", \"bar\"]"
//		}
//		serviceEventHandler.OnAdd(testServices[i])
//	}
//
//	testEndpoints := make([]*v1.Endpoints, 10)
//	endpointEventHandler := cache.EndpointsEventHandler()
//	for i := 0; i < len(testServices); i++ {
//		testEndpoints[i] = &v1.Endpoints{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:        fmt.Sprintf("%d", i),
//				Namespace:   metav1.NamespaceDefault,
//				Annotations: make(map[string]string),
//			},
//		}
//		endpointEventHandler.OnAdd(testEndpoints[i])
//	}
//
//	nodeSortAndCheckFunc := func() {
//		nodes := cache.GetNodes()
//		sort.Slice(nodes, func(i, j int) bool {
//			return strings.Compare(nodes[i].Name, nodes[j].Name) < 0
//		})
//		for i := 0; i < len(testNodes); i++ {
//			if !reflect.DeepEqual(testNodes[i], nodes[i]) {
//				t.Errorf("%d node is not equal, expect %#v to %#v", i, testNodes[i], nodes[i])
//				return
//			}
//		}
//	}
//	nodeSortAndCheckFunc()
//
//	serviceSortAndFunc := func() {
//		services := cache.GetServices()
//		sort.Slice(services, func(i, j int) bool {
//			return strings.Compare(services[i].Name, services[j].Name) < 0
//		})
//		for i := 0; i < len(testServices); i++ {
//			if !reflect.DeepEqual(testServices[i], services[i]) {
//				t.Errorf("%d service is not equal, expect %#v to %#v", i, testServices[i], services[i])
//				return
//			}
//		}
//	}
//	serviceSortAndFunc()
//
//	endpointsSortAndCheckFunc := func() {
//		endpoints := cache.GetEndpoints()
//		sort.Slice(endpoints, func(i, j int) bool {
//			return strings.Compare(endpoints[i].Name, endpoints[j].Name) < 0
//		})
//		for i := 0; i < len(testEndpoints); i++ {
//			if !reflect.DeepEqual(testEndpoints[i], endpoints[i]) {
//				t.Errorf("%d endpoint is not equal, expect %#v to %#v", i, testEndpoints[i], endpoints[i])
//				return
//			}
//		}
//	}
//	endpointsSortAndCheckFunc()
//
//	for i := 0; i < len(testNodes); i++ {
//		if i%2 == 0 {
//			testNodes[i].Labels["updated"] = "true"
//			nodeEventHandler.OnUpdate(nil, testNodes[i])
//		}
//	}
//	nonExistNode := testNodes[0].DeepCopy()
//	nonExistNode.Name = "non-exist"
//	nodeEventHandler.OnUpdate(nil, nonExistNode)
//	nodeSortAndCheckFunc()
//
//	for i := 0; i < len(testServices); i++ {
//		if i%3 == 0 {
//			testServices[i].Annotations["updated"] = "true"
//			serviceEventHandler.OnUpdate(nil, testServices[i])
//		}
//	}
//	nonExistService := testServices[0].DeepCopy()
//	nonExistService.Name = "non-exist"
//	serviceEventHandler.OnUpdate(nil, nonExistService)
//	serviceSortAndFunc()
//
//	for i := 0; i < len(testEndpoints); i++ {
//		if i%5 == 0 {
//			testEndpoints[i].Annotations["updated"] = "true"
//			endpointEventHandler.OnUpdate(nil, testEndpoints[i])
//		}
//	}
//	nonExistEndpoints := testEndpoints[0].DeepCopy()
//	nonExistEndpoints.Name = "non-exist"
//	endpointEventHandler.OnUpdate(nil, nonExistEndpoints)
//	endpointsSortAndCheckFunc()
//
//	for i := 0; i < len(testNodes); i++ {
//		nodeEventHandler.OnDelete(testNodes[i])
//	}
//	nodeEventHandler.OnDelete(nonExistNode)
//	testNodes = make([]*v1.Node, 0)
//	nodeSortAndCheckFunc()
//
//	for i := 0; i < len(testServices); i++ {
//		serviceEventHandler.OnDelete(testServices[i])
//	}
//	serviceEventHandler.OnDelete(nonExistService)
//	testServices = make([]*v1.Service, 0)
//	serviceSortAndFunc()
//
//	for i := 0; i < len(testEndpoints); i++ {
//		endpointEventHandler.OnDelete(testEndpoints[i])
//	}
//	endpointEventHandler.OnDelete(nonExistEndpoints)
//	testEndpoints = make([]*v1.Endpoints, 0)
//	endpointsSortAndCheckFunc()
//}
//
//func TestCacheServiceNotifier(t *testing.T) {
//	svcCh := make(chan watch.Event, 100)
//	endpointBroadcaster := watch.NewLongQueueBroadcaster(1000, watch.DropIfChannelFull)
//	endpointSliceChV1 := make(chan watch.Event)
//	endpointSliceChV1Beta1 := make(chan watch.Event)
//	nodeBroadcaster := watch.NewLongQueueBroadcaster(1000, watch.DropIfChannelFull)
//	supportEndpointSlice := true
//	stop := false
//	defer func() {
//		stop = true
//	}()
//
//	epsWatch := endpointBroadcaster.Watch()
//	nodeWatch := nodeBroadcaster.Watch()
//	serviceEvents := make([]watch.Event, 0)
//	go func() {
//		for !stop {
//			select {
//			case s := <-svcCh:
//				serviceEvents = append(serviceEvents, s)
//			case <-epsWatch.ResultChan():
//			case <-nodeWatch.ResultChan():
//
//			}
//		}
//	}()
//
//	cache := NewStorageCache("hostname", true, false, svcCh, endpointSliceChV1, endpointSliceChV1Beta1, endpointBroadcaster, nodeBroadcaster, supportEndpointSlice)
//
//	expectServiceSequence := make([]*v1.Service, 0)
//	testServices := make([]*v1.Service, 10)
//	serviceEventHandler := cache.ServiceEventHandler()
//	for i := 0; i < len(testServices); i++ {
//		testServices[i] = &v1.Service{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:        fmt.Sprintf("%d", i),
//				Namespace:   metav1.NamespaceDefault,
//				Annotations: make(map[string]string),
//			},
//		}
//		if i%2 == 0 {
//			testServices[i].Annotations[TopologyAnnotationsKey] = "[\"foo\", \"bar\"]"
//		}
//
//		changed := testServices[i].DeepCopy()
//		serviceEventHandler.OnAdd(changed)
//		expectServiceSequence = append(expectServiceSequence, changed)
//	}
//
//	for i := 0; i < len(testServices); i++ {
//		var changed *v1.Service
//
//		if i%3 == 0 {
//			testServices[i].Annotations["updated"] = "true"
//			changed = testServices[i].DeepCopy()
//			serviceEventHandler.OnUpdate(nil, changed)
//		} else {
//			changed = testServices[i].DeepCopy()
//			serviceEventHandler.OnDelete(changed)
//		}
//		expectServiceSequence = append(expectServiceSequence, changed)
//	}
//
//	// Drain all channel data
//	time.Sleep(time.Second)
//	if len(expectServiceSequence) != len(serviceEvents) {
//		t.Errorf("events missing, expect %d to be %d", len(expectServiceSequence), len(serviceEvents))
//		return
//	}
//
//	for i := range serviceEvents {
//		eventObject := serviceEvents[i].Object.(*v1.Service)
//		if !reflect.DeepEqual(eventObject, expectServiceSequence[i]) {
//			t.Errorf("%d expect %#v to be %#v", i, eventObject, expectServiceSequence[i])
//			return
//		}
//	}
//}

/**
func TestCacheEndpointsWithNodeChange(t *testing.T) {
	svcCh := make(chan watch.Event, 100)
	epsCh := make(chan watch.Event, 100)
	endpointSliceCh := make(chan watch.Event)
	supportEndpointSlice := true
	stop := false
	defer func() {
		stop = true
	}()

	endpointsEvents := make([]watch.Event, 0)
	go func() {
		for !stop {
			select {
			case <-svcCh:
			case e := <-epsCh:
				endpointsEvents = append(endpointsEvents, e)
			}
		}
	}()

	cache := NewStorageCache("hostname", true, false, svcCh, epsCh, endpointSliceCh, supportEndpointSlice)

	hostNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hostname",
			Labels: map[string]string{
				"region": "1",
			},
		},
	}

	testServices := []*v1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "region-1",
				Annotations: map[string]string{
					TopologyAnnotationsKey: "[\"region\"]",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "no-region",
			},
		},
	}

	testNodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					"region": "1",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-2",
				Labels: map[string]string{
					"region": "2",
				},
			},
		},
	}

	testEndpoints := []*v1.Endpoints{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "region-1",
				Namespace: metav1.NamespaceDefault,
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP:       "127.0.0.1",
							NodeName: nodeNamePtr(hostNode.Name),
						},
						{
							IP:       "127.0.0.2",
							NodeName: nodeNamePtr(testNodes[0].Name),
						},
						{
							IP:       "127.0.0.3",
							NodeName: nodeNamePtr(testNodes[1].Name),
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-region",
				Namespace: metav1.NamespaceDefault,
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP:       "127.0.2.1",
							NodeName: nodeNamePtr(hostNode.Name),
						},
						{
							IP:       "127.0.2.2",
							NodeName: nodeNamePtr(testNodes[0].Name),
						},
						{
							IP:       "127.0.2.3",
							NodeName: nodeNamePtr(testNodes[1].Name),
						},
					},
				},
			},
		},
	}

	expectedEndpoints := make([]*v1.Endpoints, len(testEndpoints))
	for i := range testEndpoints {
		expectedEndpoints[i] = testEndpoints[i].DeepCopy()
	}
	// host node has first two subnets
	// no-region has all subnets
	expectedEndpoints[0].Subsets[0].Addresses = expectedEndpoints[0].Subsets[0].Addresses[:2]
	sortEndpoints(expectedEndpoints)

	nodeEventHandler := cache.NodeEventHandler()
	nodeEventHandler.OnAdd(hostNode)
	for _, n := range testNodes {
		nodeEventHandler.OnAdd(n)
	}
	for _, s := range testServices {
		cache.ServiceEventHandler().OnAdd(s)
	}
	for _, e := range testEndpoints {
		cache.EndpointsEventHandler().OnAdd(e)
	}

	gotEndpoints := cache.GetEndpoints()
	sortEndpoints(gotEndpoints)
	for i, e := range gotEndpoints {
		if !apiequality.Semantic.DeepEqual(e, expectedEndpoints[i]) {
			t.Errorf("%d endpoints is not equal, expect %#v to be %+#v", i, e, expectedEndpoints[i])
			return
		}
	}

	// clear old events
	drainChannel(epsCh)
	endpointsEvents = make([]watch.Event, 0)
	//Change host node label
	changeHostNode := hostNode.DeepCopy()
	changeHostNode.Labels["region"] = "2"
	nodeEventHandler.OnUpdate(nil, changeHostNode)

	// host node has the first and the last subnets
	for i := range testEndpoints {
		expectedEndpoints[i] = testEndpoints[i].DeepCopy()
	}
	expectedEndpoints[0].Subsets[0].Addresses = append(testEndpoints[0].Subsets[0].Addresses[0:1],
		testEndpoints[0].Subsets[0].Addresses[2])
	sortEndpoints(expectedEndpoints)

	gotEndpoints = cache.GetEndpoints()
	sortEndpoints(gotEndpoints)
	for i, e := range gotEndpoints {
		if !apiequality.Semantic.DeepEqual(e, expectedEndpoints[i]) {
			t.Errorf("%d endpoints is not equal, expect %#v to be %+#v", i, e, expectedEndpoints[i])
			return
		}
	}

	// Drain all channel data
	time.Sleep(time.Second)
	if len(endpointsEvents) != 1 {
		t.Errorf("expect endpoints event size to be 1, but got %d", len(endpointsEvents))
		return
	}

	changedEndpoints := expectedEndpoints[1:]
	sortEndpoints(changedEndpoints)
	for i := range endpointsEvents {
		eventObject := endpointsEvents[i].Object.(*v1.Endpoints)
		if !apiequality.Semantic.DeepEqual(eventObject, changedEndpoints[i]) {
			t.Errorf("%d expect %#v to be %#v", i, eventObject, expectedEndpoints[i])
			return
		}
	}

	// clear old events
	drainChannel(epsCh)
	endpointsEvents = make([]watch.Event, 0)
	//Change non-host node label
	changeRegionNode := testNodes[1].DeepCopy()
	changeRegionNode.Labels["region"] = "3"
	nodeEventHandler.OnUpdate(nil, changeRegionNode)

	// host node has the first subnets
	for i := range testEndpoints {
		expectedEndpoints[i] = testEndpoints[i].DeepCopy()
	}
	expectedEndpoints[0].Subsets[0].Addresses = testEndpoints[0].Subsets[0].Addresses[0:1]
	sortEndpoints(expectedEndpoints)

	gotEndpoints = cache.GetEndpoints()
	sortEndpoints(gotEndpoints)
	for i, e := range gotEndpoints {
		if !apiequality.Semantic.DeepEqual(e, expectedEndpoints[i]) {
			t.Errorf("%d endpoints is not equal, expect %#v to be %+#v", i, e, expectedEndpoints[i])
			return
		}
	}

	// Drain all channel data
	time.Sleep(time.Second)
	if len(endpointsEvents) != 1 {
		t.Errorf("expect endpoints event size to be 1, but got %d", len(endpointsEvents))
		return
	}

	changedEndpoints = expectedEndpoints[1:]
	sortEndpoints(changedEndpoints)
	for i := range endpointsEvents {
		eventObject := endpointsEvents[i].Object.(*v1.Endpoints)
		if !apiequality.Semantic.DeepEqual(eventObject, changedEndpoints[i]) {
			t.Errorf("%d expect %#v to be %#v", i, eventObject, expectedEndpoints[i])
			return
		}
	}
}

*/
/**
func TestCacheEndpointsWithServiceUpdate(t *testing.T) {
	svcCh := make(chan watch.Event, 100)
	epsCh := make(chan watch.Event, 100)
	endpointSliceCh := make(chan watch.Event)
	supportEndpointSlice := true
	stop := false
	defer func() {
		stop = true
	}()

	endpointsEvents := make([]watch.Event, 0)
	go func() {
		for !stop {
			select {
			case <-svcCh:
			case e := <-epsCh:
				endpointsEvents = append(endpointsEvents, e)
			}
		}
	}()

	cache := NewStorageCache("hostname", true, false, svcCh, epsCh, endpointSliceCh, supportEndpointSlice)

	hostNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hostname",
			Labels: map[string]string{
				"region": "1",
			},
		},
	}

	testServices := []*v1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "region-1",
				Annotations: map[string]string{
					TopologyAnnotationsKey: "[\"region\"]",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   metav1.NamespaceDefault,
				Name:        "no-region",
				Annotations: make(map[string]string),
			},
		},
	}

	testNodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					"region": "1",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-2",
				Labels: map[string]string{
					"region": "2",
				},
			},
		},
	}

	testEndpoints := []*v1.Endpoints{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "region-1",
				Namespace: metav1.NamespaceDefault,
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP:       "127.0.0.1",
							NodeName: nodeNamePtr(hostNode.Name),
						},
						{
							IP:       "127.0.0.2",
							NodeName: nodeNamePtr(testNodes[0].Name),
						},
						{
							IP:       "127.0.0.3",
							NodeName: nodeNamePtr(testNodes[1].Name),
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-region",
				Namespace: metav1.NamespaceDefault,
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP:       "127.0.2.1",
							NodeName: nodeNamePtr(hostNode.Name),
						},
						{
							IP:       "127.0.2.2",
							NodeName: nodeNamePtr(testNodes[0].Name),
						},
						{
							IP:       "127.0.2.3",
							NodeName: nodeNamePtr(testNodes[1].Name),
						},
					},
				},
			},
		},
	}

	expectedEndpoints := make([]*v1.Endpoints, len(testEndpoints))
	for i := range testEndpoints {
		expectedEndpoints[i] = testEndpoints[i].DeepCopy()
	}
	// host node has first two subnets
	// no-region has all subnets
	expectedEndpoints[0].Subsets[0].Addresses = expectedEndpoints[0].Subsets[0].Addresses[:2]
	sortEndpoints(expectedEndpoints)

	cache.NodeEventHandler().OnAdd(hostNode)
	for _, n := range testNodes {
		cache.NodeEventHandler().OnAdd(n)
	}
	serviceEventHandler := cache.ServiceEventHandler()
	for _, s := range testServices {
		serviceEventHandler.OnAdd(s)
	}
	for _, e := range testEndpoints {
		cache.EndpointsEventHandler().OnAdd(e)
	}

	gotEndpoints := cache.GetEndpoints()
	sortEndpoints(gotEndpoints)
	for i, e := range gotEndpoints {
		if !apiequality.Semantic.DeepEqual(e, expectedEndpoints[i]) {
			t.Errorf("%d endpoints is not equal, expect %#v to be %+#v", i, e, expectedEndpoints[i])
			return
		}
	}

	// clear old events
	drainChannel(epsCh)
	endpointsEvents = make([]watch.Event, 0)
	// change non-relative service annotations
	changeService := testServices[1].DeepCopy()
	changeService.Annotations["non-relative"] = "true"
	serviceEventHandler.OnUpdate(nil, changeService)

	gotEndpoints = cache.GetEndpoints()
	sortEndpoints(gotEndpoints)
	for i, e := range gotEndpoints {
		if !apiequality.Semantic.DeepEqual(e, expectedEndpoints[i]) {
			t.Errorf("%d endpoints is not equal, expect %#v to be %+#v", i, e, expectedEndpoints[i])
			return
		}
	}

	// clear old events
	drainChannel(epsCh)
	endpointsEvents = make([]watch.Event, 0)
	// no-region service has topology key
	changeService = testServices[1].DeepCopy()
	changeService.Annotations[TopologyAnnotationsKey] = "[\"region\"]"
	serviceEventHandler.OnUpdate(nil, changeService)

	for i := range testEndpoints {
		expectedEndpoints[i] = testEndpoints[i].DeepCopy()
		expectedEndpoints[i].Subsets[0].Addresses = expectedEndpoints[i].Subsets[0].Addresses[:2]
	}
	sortEndpoints(expectedEndpoints)

	gotEndpoints = cache.GetEndpoints()
	sortEndpoints(gotEndpoints)
	for i, e := range gotEndpoints {
		if !apiequality.Semantic.DeepEqual(e, expectedEndpoints[i]) {
			t.Errorf("%d endpoints is not equal, expect %#v to be %+#v", i, e, expectedEndpoints[i])
			return
		}
	}

	// Drain all channel data
	time.Sleep(time.Second)
	if len(endpointsEvents) != 1 {
		t.Errorf("expect endpoints event size to be 1, but got %d", len(endpointsEvents))
		return
	}

	changedEndpoints := expectedEndpoints[0:1]
	sortEndpoints(changedEndpoints)
	for i := range endpointsEvents {
		eventObject := endpointsEvents[i].Object.(*v1.Endpoints)
		if !apiequality.Semantic.DeepEqual(eventObject, changedEndpoints[i]) {
			t.Errorf("%d expect %#v to be %#v", i, eventObject, expectedEndpoints[i])
			return
		}
	}
}

*/

func nodeNamePtr(nodeName string) *string {
	n := nodeName
	return &n
}

func sortEndpoints(arr []*v1.Endpoints) {
	sort.Slice(arr, func(i, j int) bool {
		strCmpResult := strings.Compare(arr[i].Name, arr[j].Name)
		return strCmpResult < 0
	})
}

func drainChannel(ch chan watch.Event) {
	drained := false
	for !drained {
		select {
		case <-ch:
		case <-time.After(time.Millisecond * 10):
			drained = true
		}
	}
}

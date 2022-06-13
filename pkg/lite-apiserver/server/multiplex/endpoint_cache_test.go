/*
Copyright 2014 The Kubernetes Authors.

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

package multiplex

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiwatch "k8s.io/apimachinery/pkg/watch"
	utiltesting "k8s.io/client-go/util/testing"
)

var emptyNodeName = "nil"

func makeTestServer(t *testing.T) (*httptest.Server, *utiltesting.FakeHandler) {
	fakeEndpointsHandler := utiltesting.FakeHandler{
		StatusCode:   http.StatusOK,
		ResponseBody: runtime.EncodeOrDie(clientscheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), &v1.Endpoints{}),
	}
	mux := http.NewServeMux()
	mux.Handle("/api/v1/endpoints", &fakeEndpointsHandler)
	mux.Handle("/api/v1/endpoints/", &fakeEndpointsHandler)
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		t.Errorf("unexpected request: %v", req.RequestURI)
		http.Error(res, "", http.StatusNotFound)
	})
	return httptest.NewServer(mux), &fakeEndpointsHandler
}

func newSharedInformerFactory(url string) informers.SharedInformerFactory {
	client := clientset.NewForConfigOrDie(&restclient.Config{Host: url, ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	return informerFactory
}

func TestEndpointMuxWatch(t *testing.T) {
	testServer, _ := makeTestServer(t)

	mux, err := NewEndpointMux("testnode", newSharedInformerFactory(testServer.URL))
	if err != nil {
		t.Fatalf("NewEndpointMux returned unexpected error %v", err)
	}
	epmux := mux.(*endpointMux)

	storecase := []*v1.Endpoints{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "store1",
				Namespace:       "defalut",
				ResourceVersion: "1",
			},
			Subsets: []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{IP: "1.2.3.4", NodeName: &emptyNodeName}},
				Ports:     []v1.EndpointPort{{Port: 1000, Protocol: "TCP"}},
			}},
		},
	}
	epmux.epInformer.GetStore().Add(storecase[0])

	watchcase := []apiwatch.Event{
		{
			Type: apiwatch.Added,
			Object: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					Namespace:       "defalut",
					ResourceVersion: "1",
				},
				Subsets: []v1.EndpointSubset{{
					Addresses: []v1.EndpointAddress{{IP: "6.7.8.9", NodeName: &emptyNodeName}},
					Ports:     []v1.EndpointPort{{Port: 1000, Protocol: "TCP"}},
				}},
			}},
		{
			Type: apiwatch.Added,
			Object: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo1",
					Namespace:       "defalut",
					ResourceVersion: "1",
				},
				Subsets: []v1.EndpointSubset{{
					Addresses: []v1.EndpointAddress{{IP: "1.1.1.1", NodeName: &emptyNodeName}},
					Ports:     []v1.EndpointPort{{Port: 1000, Protocol: "TCP"}},
				}},
			}},
		{
			Type: apiwatch.Modified,
			Object: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo2",
					Namespace:       "defalut",
					ResourceVersion: "1",
				},
				Subsets: []v1.EndpointSubset{{
					Addresses: []v1.EndpointAddress{{IP: "2.2.2.2", NodeName: &emptyNodeName}},
					Ports:     []v1.EndpointPort{{Port: 1000, Protocol: "TCP"}},
				}},
			}},
		{
			Type: apiwatch.Deleted,
			Object: &v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo3",
					Namespace:       "defalut",
					ResourceVersion: "1",
				},
				Subsets: []v1.EndpointSubset{{
					Addresses: []v1.EndpointAddress{{IP: "3.3.3.3", NodeName: &emptyNodeName}},
					Ports:     []v1.EndpointPort{{Port: 1000, Protocol: "TCP"}},
				}},
			}},
	}
	table := []apiwatch.Event{
		{
			Type:   apiwatch.Added,
			Object: storecase[0],
		},
	}
	table = append(table, watchcase...)
	const testWatchers = 2
	wg := sync.WaitGroup{}
	wg.Add(testWatchers)
	for i := 0; i < testWatchers; i++ {
		w, err := mux.Watch(false, "")
		if err != nil {
			t.Fatalf("mux.Watch() returned unexpected error %v", err)
		}
		go func(watcher int, w apiwatch.Interface) {
			tableLine := 0
			for {
				// the event is from store first
				event, ok := <-w.ResultChan()
				if !ok {
					break
				}
				if e, a := table[tableLine], event; !reflect.DeepEqual(e, a) {
					t.Errorf("Watcher %v, line %v: Expected (%v, %#v), got (%v, %#v)",
						watcher, tableLine, e.Type, e.Object, a.Type, a.Object)
				} else {
					t.Logf("watcher %d Got (%v, %#v)", watcher, event.Type, event.Object)
				}
				tableLine++
			}
			wg.Done()
		}(i, w)
	}

	epmux.OnAdd(table[1].Object)
	epmux.OnAdd(table[2].Object)
	epmux.OnUpdate(nil, table[3].Object)
	epmux.OnDelete(table[4].Object)
	epmux.broadcaster.Shutdown()

	wg.Wait()
}

func TestEndpointMuxListObjects(t *testing.T) {
	testServer, _ := makeTestServer(t)

	mux, err := NewEndpointMux("testnode", newSharedInformerFactory(testServer.URL))
	if err != nil {
		t.Fatalf("NewEndpointMux returned unexpected error %v", err)
	}
	epmux := mux.(*endpointMux)

	storecase := []*v1.Endpoints{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "store1",
				Namespace:       "defalut",
				ResourceVersion: "1",
				Labels:          map[string]string{"foo": "bar"},
			},
			Subsets: []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{IP: "1.2.3.4", NodeName: &emptyNodeName}},
				Ports:     []v1.EndpointPort{{Port: 1000, Protocol: "TCP"}},
			}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "store2",
				Namespace:       "defalut",
				ResourceVersion: "1",
				Labels:          map[string]string{"foo1": "bar1"},
			},
			Subsets: []v1.EndpointSubset{{
				Addresses: []v1.EndpointAddress{{IP: "2.2.3.4", NodeName: &emptyNodeName}},
				Ports:     []v1.EndpointPort{{Port: 1000, Protocol: "TCP"}},
			}},
		},
	}
	for _, e := range storecase {
		epmux.epInformer.GetStore().Add(e)
	}
	var eps1 []*v1.Endpoints
	epmux.ListObjects(labels.Everything(), func(m interface{}) {
		eps1 = append(eps1, m.(*v1.Endpoints))
	})
	if len(eps1) != 2 {
		t.Errorf("ListObject count error Expected  (%d), got (%d)", len(storecase), len(eps1))
	}

	var eps2 []*v1.Endpoints
	label, _ := labels.Parse("foo=bar")
	epmux.ListObjects(label, func(m interface{}) {
		eps2 = append(eps2, m.(*v1.Endpoints))
	})
	if len(eps2) != 1 {
		t.Errorf("ListObject count error Expected  (%d), got (%d)", 1, len(eps2))
	}

}

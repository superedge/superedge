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

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/cmd/application-grid-wrapper/app/options"
	"github.com/superedge/superedge/pkg/edge-health/data"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/wait"
	informcorev1 "k8s.io/client-go/informers/core/v1"
	"net/http"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/application-grid-wrapper/server/apis"
	"github.com/superedge/superedge/pkg/application-grid-wrapper/storage"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	apischema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/restmapper"
)

const (
	NODELABELS_INDEXER = "nodeLabels"
)

type interceptorServer struct {
	restConfig                        *rest.Config
	cache                             storage.Cache
	serviceWatchCh                    <-chan watch.Event
	endpointsBoradcaster              *watch.Broadcaster
	endpointSliceV1WatchCh            <-chan watch.Event
	endpointSliceV1Beta1WatchCh       <-chan watch.Event
	nodeBoradcaster                   *watch.Broadcaster
	mediaSerializer                   []runtime.SerializerInfo
	serviceAutonomyEnhancementAddress string
	supportEndpointSlice              bool
	nodeIndexer                       cache.Indexer
}

func NodeLabelsFunc(obj interface{}) ([]string, error) {
	node, ok := obj.(*v1.Node)
	if ok {
		labels := []string{}
		for k, v := range node.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}
		return labels, nil
	}
	return []string{}, nil
}
func NewInterceptorServer(kubeconfig string, hostName string, wrapperInCluster bool, channelSize int, serviceAutonomyEnhancement options.ServiceAutonomyEnhancementOptions, supportEndpointSlice bool) *interceptorServer {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Errorf("can't build rest config, %v", err)
		return nil
	}

	serviceCh := make(chan watch.Event, channelSize)
	endpointBroadcaster := watch.NewLongQueueBroadcaster(channelSize, watch.DropIfChannelFull)
	endpointSliceV1Ch := make(chan watch.Event, channelSize)
	endpointSliceV1Beta1Ch := make(chan watch.Event, channelSize)
	nodeBroadcaster := watch.NewLongQueueBroadcaster(channelSize, watch.DropIfChannelFull)

	server := &interceptorServer{
		restConfig:                        restConfig,
		cache:                             storage.NewStorageCache(hostName, wrapperInCluster, serviceAutonomyEnhancement.Enabled, serviceCh, endpointSliceV1Ch, endpointSliceV1Beta1Ch, endpointBroadcaster, nodeBroadcaster, supportEndpointSlice),
		serviceWatchCh:                    serviceCh,
		endpointsBoradcaster:              endpointBroadcaster,
		endpointSliceV1WatchCh:            endpointSliceV1Ch,
		endpointSliceV1Beta1WatchCh:       endpointSliceV1Beta1Ch,
		nodeBoradcaster:                   nodeBroadcaster,
		mediaSerializer:                   scheme.Codecs.SupportedMediaTypes(),
		serviceAutonomyEnhancementAddress: serviceAutonomyEnhancement.NeighborStatusSvc,
		supportEndpointSlice:              supportEndpointSlice,
	}

	return server
}

func (s *interceptorServer) Run(debug bool, bindAddress string, insecure bool, caFile, certFile, keyFile string, serviceAutonomyEnhancement options.ServiceAutonomyEnhancementOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/* cache
	 */
	if err := s.setupInformers(ctx.Done()); err != nil {
		return err
	}

	klog.Infof("Start to run GetLocalInfo client")

	if serviceAutonomyEnhancement.Enabled {
		go wait.Until(s.NodeStatusAcquisition, time.Duration(serviceAutonomyEnhancement.UpdateInterval)*time.Second, ctx.Done())
	}

	klog.Infof("Start to run interceptor server")
	/* filter
	 */
	server := &http.Server{Addr: bindAddress, Handler: s.buildFilterChains(debug)}

	if insecure {
		return server.ListenAndServe()
	}

	tlsConfig := &tls.Config{}
	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(caFile)
	if err != nil {
		klog.Errorf("can't read ca file %s, %v", caFile, err)
		return nil
	}
	pool.AppendCertsFromPEM(caCrt)
	tlsConfig.RootCAs = pool

	if len(certFile) != 0 && len(keyFile) != 0 {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			klog.Errorf("can't load certificate pair %s %s, %v", certFile, keyFile, err)
			return nil
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}

	server.TLSConfig = tlsConfig
	return server.ListenAndServeTLS("", "")
}

func (s *interceptorServer) setupInformers(stop <-chan struct{}) error {
	klog.Infof("Start to run service and endpoints informers")
	noProxyName, err := labels.NewRequirement(apis.LabelServiceProxyName, selection.DoesNotExist, nil)
	if err != nil {
		klog.Errorf("can't parse proxy label, %v", err)
		return err
	}

	noHeadlessEndpoints, err := labels.NewRequirement(v1.IsHeadlessService, selection.DoesNotExist, nil)
	if err != nil {
		klog.Errorf("can't parse headless label, %v", err)
		return err
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*noProxyName, *noHeadlessEndpoints)

	resyncPeriod := time.Minute * 5
	client := kubernetes.NewForConfigOrDie(s.restConfig)
	nodeInformerFactory := informers.NewSharedInformerFactory(client, resyncPeriod)
	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, resyncPeriod,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector.String()
		}))

	nodeInformer := nodeInformerFactory.InformerFor(&v1.Node{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
		return informcorev1.NewNodeInformer(k, duration, cache.Indexers{NODELABELS_INDEXER: NodeLabelsFunc})
	})
	serviceInformer := informerFactory.Core().V1().Services().Informer()
	endpointsInformer := informerFactory.Core().V1().Endpoints().Informer()

	nodeInformer.AddEventHandlerWithResyncPeriod(s.cache.NodeEventHandler(), resyncPeriod)
	serviceInformer.AddEventHandlerWithResyncPeriod(s.cache.ServiceEventHandler(), resyncPeriod)
	endpointsInformer.AddEventHandlerWithResyncPeriod(s.cache.EndpointsEventHandler(), resyncPeriod)

	go nodeInformer.Run(stop)
	go serviceInformer.Run(stop)
	go endpointsInformer.Run(stop)

	if !cache.WaitForNamedCacheSync("node", stop,
		nodeInformer.HasSynced,
		serviceInformer.HasSynced,
		endpointsInformer.HasSynced) {
		return fmt.Errorf("can't sync informers")
	}

	s.nodeIndexer = nodeInformer.GetIndexer()

	restMapperRes, err := restmapper.GetAPIGroupResources(client.Discovery())
	if err != nil {
		epsv1, err := client.DiscoveryV1().EndpointSlices("").List(context.Background(), metav1.ListOptions{})
		if err == nil && len(epsv1.Items) != 0 {
			klog.Info("start v1.EndpointSlices informer")
			endpointSliceV1Informer := informerFactory.Discovery().V1().EndpointSlices().Informer()
			endpointSliceV1Informer.AddEventHandlerWithResyncPeriod(s.cache.EndpointSliceV1EventHandler(), resyncPeriod)
			go endpointSliceV1Informer.Run(stop)
			if !cache.WaitForNamedCacheSync("node", stop, endpointSliceV1Informer.HasSynced) {
				return fmt.Errorf("can't sync endpointslice informers")
			}
		} else {
			epsv1beta1, err := client.DiscoveryV1beta1().EndpointSlices("").List(context.Background(), metav1.ListOptions{})
			if err == nil && len(epsv1beta1.Items) != 0 {
				klog.Info("start v1beta1.EndpointSlices informer")
				endpointSliceV1Beta1Informer := informerFactory.Discovery().V1beta1().EndpointSlices().Informer()
				endpointSliceV1Beta1Informer.AddEventHandlerWithResyncPeriod(s.cache.EndpointSliceV1Beta1EventHandler(), resyncPeriod)
				go endpointSliceV1Beta1Informer.Run(stop)
				if !cache.WaitForNamedCacheSync("node", stop, endpointSliceV1Beta1Informer.HasSynced) {
					return fmt.Errorf("can't sync endpointslice informers")
				}
			}
		}
		return err
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(restMapperRes)
	_, err = restMapper.RESTMapping(apischema.GroupKind{
		Group: discoveryv1.SchemeGroupVersion.Group,
		Kind:  "EndpointSlice",
	}, discoveryv1.SchemeGroupVersion.Version)
	if err == nil {
		klog.Info("mapper v1.EndpointSlices")
		endpointSliceV1Informer := informerFactory.Discovery().V1().EndpointSlices().Informer()
		endpointSliceV1Informer.AddEventHandlerWithResyncPeriod(s.cache.EndpointSliceV1EventHandler(), resyncPeriod)
		go endpointSliceV1Informer.Run(stop)
		if !cache.WaitForNamedCacheSync("node", stop, endpointSliceV1Informer.HasSynced) {
			return fmt.Errorf("can't sync endpointslice informers")
		}
	} else {
		if _, ok := err.(*meta.NoKindMatchError); ok {
			_, err = restMapper.RESTMapping(apischema.GroupKind{
				Group: discoveryv1beta1.SchemeGroupVersion.Group,
				Kind:  "EndpointSlice",
			}, discoveryv1beta1.SchemeGroupVersion.Version)
			if err == nil {
				klog.Info("mapper v1beta1.EndpointSlices")
				endpointSliceV1Beta1Informer := informerFactory.Discovery().V1beta1().EndpointSlices().Informer()
				endpointSliceV1Beta1Informer.AddEventHandlerWithResyncPeriod(s.cache.EndpointSliceV1Beta1EventHandler(), resyncPeriod)
				go endpointSliceV1Beta1Informer.Run(stop)
				if !cache.WaitForNamedCacheSync("node", stop, endpointSliceV1Beta1Informer.HasSynced) {
					return fmt.Errorf("can't sync endpointslice informers")
				}
			} else {
				if _, ok := err.(*meta.NoKindMatchError); !ok {
					return err
				}
			}
		} else {
			return err
		}
	}
	return nil
}

func (s *interceptorServer) buildFilterChains(debug bool) http.Handler {
	upstream, err := Newupstream(s.restConfig)
	if err != nil {
		klog.Errorf("Fialed to get upstreamHandler, error: %v")
		klog.Fatal(err)
	}
	handler := http.Handler(upstream)
	handler = s.interceptIngressEndpointsRequest(handler)
	handler = s.interceptIngressRequest(handler)
	handler = s.interceptEndpointSliceV1Beta1Request(handler)
	handler = s.interceptEndpointSliceV1Request(handler)
	handler = s.interceptEndpointsRequest(handler)
	handler = s.interceptServiceRequest(handler)
	handler = s.interceptEventRequest(handler)
	handler = s.interceptNodeRequest(handler)
	handler = s.logger(handler)

	if debug {
		handler = s.debugger(handler)
	}

	return handler
}

func (s *interceptorServer) NodeStatusAcquisition() {
	client := http.Client{Timeout: 5 * time.Second}
	klog.V(4).Infof("serviceAutonomyEnhancementAddress is %s", s.serviceAutonomyEnhancementAddress)
	req, err := http.NewRequest("GET", s.serviceAutonomyEnhancementAddress, nil)
	if err != nil {
		s.cache.ClearLocalNodeInfo()
		klog.Errorf("Get local node info: new request err: %v", err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		s.cache.ClearLocalNodeInfo()
		klog.Errorf("Get local node info: do request err: %v", err)
		return
	}
	defer func() {
		if res != nil {
			res.Body.Close()
		}
	}()

	if res.StatusCode != http.StatusOK {
		klog.Errorf("Get local node info: httpResponse.StatusCode!=200, is %d", res.StatusCode)
		s.cache.ClearLocalNodeInfo()
		return
	}

	var localNodeInfo map[string]data.ResultDetail
	if err := json.NewDecoder(res.Body).Decode(&localNodeInfo); err != nil {
		klog.Errorf("Get local node info: Decode err: %v", err)
		s.cache.ClearLocalNodeInfo()
		return
	} else {
		s.cache.SetLocalNodeInfo(localNodeInfo)
	}
	return
}

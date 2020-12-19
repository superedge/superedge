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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	v1 "k8s.io/api/core/v1"
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
	"k8s.io/klog"

	"superedge/pkg/application-grid-wrapper/server/apis"
	"superedge/pkg/application-grid-wrapper/storage"
)

type interceptorServer struct {
	restConfig       *rest.Config
	cache            storage.Cache
	serviceWatchCh   <-chan watch.Event
	endpointsWatchCh <-chan watch.Event
	mediaSerializer  []runtime.SerializerInfo
}

func NewInterceptorServer(kubeconfig string, hostName string, wrapperInCluster bool, channelSize int) *interceptorServer {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Errorf("can't build rest config, %v", err)
		return nil
	}

	serviceCh := make(chan watch.Event, channelSize)
	endpointsCh := make(chan watch.Event, channelSize)

	server := &interceptorServer{
		restConfig:       restConfig,
		cache:            storage.NewStorageCache(hostName, wrapperInCluster, serviceCh, endpointsCh),
		serviceWatchCh:   serviceCh,
		endpointsWatchCh: endpointsCh,
		mediaSerializer:  scheme.Codecs.SupportedMediaTypes(),
	}

	return server
}

func (s *interceptorServer) Run(debug bool, bindAddress string, insecure bool, caFile, certFile, keyFile string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/* cache
	 */
	if err := s.setupInformers(ctx.Done()); err != nil {
		return err
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
	klog.Infof("Start to run service and endpoints infromers")
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

	nodeInformer := nodeInformerFactory.Core().V1().Nodes().Informer()
	serviceInformer := informerFactory.Core().V1().Services().Informer()
	endpointsInformer := informerFactory.Core().V1().Endpoints().Informer()

	/*
	 */
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

	return nil
}

func (s *interceptorServer) buildFilterChains(debug bool) http.Handler {
	handler := http.Handler(http.NewServeMux())

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

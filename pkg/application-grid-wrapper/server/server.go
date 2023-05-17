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
	"io/ioutil"
	"net"
	"net/http"
	"time"

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	apischema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	informcorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/cmd/application-grid-wrapper/app/options"
	"github.com/superedge/superedge/pkg/application-grid-wrapper/server/apis"
	"github.com/superedge/superedge/pkg/application-grid-wrapper/storage"
	"github.com/superedge/superedge/pkg/edge-health/data"
	sitev1alpha2 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	siteconstant "github.com/superedge/superedge/pkg/site-manager/constant"
	"github.com/superedge/superedge/pkg/site-manager/controller/unitcluster"
	crdclientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
)

const (
	NODELABELS_INDEXER = "nodeLabels"
)

type interceptorServer struct {
	restConfig                           *rest.Config
	cache                                storage.Cache
	serviceWatchBroadcaster              *watch.Broadcaster
	endpointsBroadcaster                 *watch.Broadcaster
	endpointSliceV1WatchBroadcaster      *watch.Broadcaster
	endpointSliceV1Beta1WatchBroadcaster *watch.Broadcaster
	nodeBoradcaster                      *watch.Broadcaster
	k3sServiceInformer                   cache.SharedIndexInformer
	k3sNodeInformer                      cache.SharedIndexInformer
	k3sEndpointInformer                  cache.SharedIndexInformer
	k3sEndpointSliceV1Informer           cache.SharedIndexInformer
	k3sEndpointSliceV1Beta1Informer      cache.SharedIndexInformer
	mediaSerializer                      []runtime.SerializerInfo
	serviceAutonomyEnhancementAddress    string
	supportEndpointSlice                 bool
	nodeIndexer                          cache.Indexer
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
func NewInterceptorServer(restConfig *rest.Config, hostName string, wrapperInCluster bool, channelSize int, serviceAutonomyEnhancement options.ServiceAutonomyEnhancementOptions, supportEndpointSlice bool) *interceptorServer {
	serviceBroadcaster := watch.NewLongQueueBroadcaster(channelSize, watch.DropIfChannelFull)
	endpointBroadcaster := watch.NewLongQueueBroadcaster(channelSize, watch.DropIfChannelFull)
	endpointSliceV1Broadcaster := watch.NewLongQueueBroadcaster(channelSize, watch.DropIfChannelFull)
	endpointSliceV1Beta1Broadcaster := watch.NewLongQueueBroadcaster(channelSize, watch.DropIfChannelFull)
	nodeBroadcaster := watch.NewLongQueueBroadcaster(channelSize, watch.DropIfChannelFull)

	server := &interceptorServer{
		restConfig:                           restConfig,
		cache:                                storage.NewStorageCache(hostName, wrapperInCluster, serviceAutonomyEnhancement.Enabled, serviceBroadcaster, endpointSliceV1Broadcaster, endpointSliceV1Beta1Broadcaster, endpointBroadcaster, nodeBroadcaster, supportEndpointSlice),
		serviceWatchBroadcaster:              serviceBroadcaster,
		endpointsBroadcaster:                 endpointBroadcaster,
		endpointSliceV1WatchBroadcaster:      endpointSliceV1Broadcaster,
		endpointSliceV1Beta1WatchBroadcaster: endpointSliceV1Beta1Broadcaster,
		nodeBoradcaster:                      nodeBroadcaster,
		mediaSerializer:                      scheme.Codecs.SupportedMediaTypes(),
		serviceAutonomyEnhancementAddress:    serviceAutonomyEnhancement.NeighborStatusSvc,
		supportEndpointSlice:                 supportEndpointSlice,
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
	} else {
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
	}

	crdClient := crdclientset.NewForConfigOrDie(s.restConfig)
	nodeLister := listerscorev1.NewNodeLister(nodeInformer.GetIndexer())
	setupK3sInformer := func(kubeconfig, nodeunit string, k3sService *v1.Service, k3sStop <-chan struct{}) error {
		k3sRestConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
		if err != nil {
			klog.Errorf("Failed to get k3s kubeclient restConfig, error: %v", err)
			return err
		}

		k3sRestConfig.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			eps, err := listersv1.NewEndpointsLister(endpointsInformer.GetIndexer()).Endpoints(k3sService.Namespace).Get(k3sService.Name)
			if err != nil {
				return nil, err
			}
			if len(eps.Subsets) == 0 {
				return nil, fmt.Errorf("all replicas of apiserver are unavailable, service: %s", fmt.Sprintf("%s.%s", k3sService.Name, k3sService.Namespace))
			}
			for _, ep := range eps.Subsets[0].Addresses {
				conn, err := net.Dial("tcp", fmt.Sprintf("%s:6443", ep.IP))
				if err == nil {
					return conn, nil
				}
			}
			return nil, fmt.Errorf("all replicas of apiserver are unavailable, service: %s", fmt.Sprintf("%s.%s", k3sService.Name, k3sService.Namespace))
		}
		k3sClient := kubernetes.NewForConfigOrDie(k3sRestConfig)

		if k3sClient != nil {
			k3sNodeInformerFactory := informers.NewSharedInformerFactory(k3sClient, resyncPeriod)
			k3sInformerFactory := informers.NewSharedInformerFactoryWithOptions(k3sClient, resyncPeriod,
				informers.WithTweakListOptions(func(options *metav1.ListOptions) {
					options.LabelSelector = labelSelector.String()
				}))

			k3sNodeInformer := k3sNodeInformerFactory.InformerFor(&v1.Node{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
				return informcorev1.NewNodeInformer(k, duration, cache.Indexers{NODELABELS_INDEXER: NodeLabelsFunc})
			})
			k3sServiceInformer := k3sInformerFactory.Core().V1().Services().Informer()
			k3sEndpointsInformer := k3sInformerFactory.Core().V1().Endpoints().Informer()

			// node eventHandler
			k3sNodeInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					s.nodeBoradcaster.ActionOrDrop(watch.Added, obj.(*v1.Node))
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					updateNode := newObj.(*v1.Node)
					node, nodeErr := nodeLister.Get(updateNode.Name)
					if apierrors.IsNotFound(nodeErr) {
						deleteErr := k3sClient.CoreV1().Nodes().Delete(context.TODO(), updateNode.Name, metav1.DeleteOptions{})
						if err != nil {
							klog.Error(deleteErr)
						}
					} else {
						if v, ok := node.Labels[nodeunit]; ok {
							if v == siteconstant.NodeUnitSuperedge {
								s.nodeBoradcaster.ActionOrDrop(watch.Modified, updateNode)
								return
							}
						}
						deleteErr := k3sClient.CoreV1().Nodes().Delete(context.TODO(), updateNode.Name, metav1.DeleteOptions{})
						if err != nil {
							klog.Error(deleteErr)
						}
					}
				},
				DeleteFunc: func(obj interface{}) {
					s.nodeBoradcaster.ActionOrDrop(watch.Deleted, obj.(*v1.Node))
				},
			}, resyncPeriod)
			// node errorEventHandler
			k3sNodeInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
				for _, v := range k3sNodeInformer.GetStore().List() {
					k3sNodeInformer.GetStore().Delete(v)
					s.nodeBoradcaster.ActionOrDrop(watch.Deleted, v.(*v1.Node))
				}
			})

			// service eventHandler
			k3sServiceInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					svc := obj.(*v1.Service)
					k3sSvc := svc.DeepCopy()
					k3sSvc.Name = fmt.Sprintf("k3s-%s", svc.Name)
					s.serviceWatchBroadcaster.ActionOrDrop(watch.Added, k3sSvc)
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					svc := newObj.(*v1.Service)
					k3sSvc := svc.DeepCopy()
					k3sSvc.Name = fmt.Sprintf("k3s-%s", svc.Name)
					s.serviceWatchBroadcaster.ActionOrDrop(watch.Modified, k3sSvc)
				},
				DeleteFunc: func(obj interface{}) {
					svc := obj.(*v1.Service)
					k3sSvc := svc.DeepCopy()
					k3sSvc.Name = fmt.Sprintf("k3s-%s", svc.Name)
					s.serviceWatchBroadcaster.ActionOrDrop(watch.Deleted, k3sSvc)
				},
			}, resyncPeriod)

			// service errorEventHandler
			k3sServiceInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
				for _, v := range k3sServiceInformer.GetStore().List() {
					k3sServiceInformer.GetStore().Delete(v)
					svc := v.(*v1.Service)
					k3sSvc := svc.DeepCopy()
					k3sSvc.Name = fmt.Sprintf("k3s-%s", svc.Name)
					s.serviceWatchBroadcaster.ActionOrDrop(watch.Deleted, k3sSvc)
				}
			})

			// endpoint eventHandler
			k3sEndpointsInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					eps := obj.(*v1.Endpoints)
					k3sEps := eps.DeepCopy()
					k3sEps.Name = fmt.Sprintf("k3s-%s", eps.Name)
					s.endpointsBroadcaster.ActionOrDrop(watch.Added, k3sEps)
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					eps := newObj.(*v1.Endpoints)
					k3sEps := eps.DeepCopy()
					k3sEps.Name = fmt.Sprintf("k3s-%s", eps.Name)
					s.endpointsBroadcaster.ActionOrDrop(watch.Modified, k3sEps)
				},
				DeleteFunc: func(obj interface{}) {
					eps := obj.(*v1.Endpoints)
					k3sEps := eps.DeepCopy()
					k3sEps.Name = fmt.Sprintf("k3s-%s", eps.Name)
					s.endpointsBroadcaster.ActionOrDrop(watch.Deleted, k3sEps)
				},
			}, resyncPeriod)
			// endpoint errEventHandler
			k3sEndpointsInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
				for _, v := range k3sEndpointsInformer.GetStore().List() {
					k3sEndpointsInformer.GetStore().Delete(v)

					eps := v.(*v1.Endpoints)
					k3sEps := eps.DeepCopy()
					k3sEps.Name = fmt.Sprintf("k3s-%s", eps.Name)
					s.endpointsBroadcaster.ActionOrDrop(watch.Deleted, v.(*v1.Endpoints))
				}
			})
			go k3sNodeInformer.Run(k3sStop)
			go k3sServiceInformer.Run(k3sStop)
			go k3sEndpointsInformer.Run(k3sStop)
			if !cache.WaitForNamedCacheSync("k3s", k3sStop,
				k3sNodeInformer.HasSynced,
				k3sServiceInformer.HasSynced,
				k3sEndpointsInformer.HasSynced) {
				return fmt.Errorf("can't sync informers")
			}

			s.k3sNodeInformer = k3sNodeInformer
			s.k3sServiceInformer = k3sServiceInformer
			s.k3sEndpointInformer = k3sEndpointsInformer

			setupEndpointSlicesV1Informer := func() error {
				klog.Info("k3s start v1.EndpointSlices informer")
				k3sEndpointSliceV1Informer := k3sInformerFactory.Discovery().V1().EndpointSlices().Informer()

				// eventHandler
				k3sEndpointSliceV1Informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						k3sEpsV1 := obj.(*discoveryv1.EndpointSlice)
						superedgeEpsV1 := k3sEpsV1.DeepCopy()
						superedgeEpsV1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1.Name)
						superedgeEpsV1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1.Labels["kubernetes.io/service-name"])
						s.endpointSliceV1WatchBroadcaster.ActionOrDrop(watch.Added, superedgeEpsV1)
					},
					UpdateFunc: func(oldObj, newObj interface{}) {
						k3sEpsV1 := newObj.(*discoveryv1.EndpointSlice)
						superedgeEpsV1 := k3sEpsV1.DeepCopy()
						superedgeEpsV1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1.Name)
						superedgeEpsV1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1.Labels["kubernetes.io/service-name"])
						s.endpointSliceV1WatchBroadcaster.ActionOrDrop(watch.Modified, superedgeEpsV1)
					},
					DeleteFunc: func(obj interface{}) {
						k3sEpsV1 := obj.(*discoveryv1.EndpointSlice)
						superedgeEpsV1 := k3sEpsV1.DeepCopy()
						superedgeEpsV1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1.Name)
						superedgeEpsV1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1.Labels["kubernetes.io/service-name"])
						s.endpointSliceV1WatchBroadcaster.ActionOrDrop(watch.Deleted, superedgeEpsV1)
					},
				}, resyncPeriod)

				// errorEventHandler
				k3sEndpointSliceV1Informer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
					for _, v := range k3sEndpointSliceV1Informer.GetStore().List() {
						k3sEndpointSliceV1Informer.GetStore().Delete(v)

						eps := v.(*discoveryv1.EndpointSlice)
						k3sEps := eps.DeepCopy()
						k3sEps.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", eps.Labels["kubernetes.io/service-name"])
						k3sEps.Name = fmt.Sprintf("k3s-%s", eps.Name)
						s.endpointSliceV1WatchBroadcaster.ActionOrDrop(watch.Deleted, k3sEps)
					}
				})
				go k3sEndpointSliceV1Informer.Run(k3sStop)
				if !cache.WaitForNamedCacheSync("k3s", k3sStop, k3sEndpointSliceV1Informer.HasSynced) {
					return fmt.Errorf("can't sync endpointslice informers")
				}
				s.k3sEndpointSliceV1Informer = k3sEndpointSliceV1Informer
				return nil
			}

			setupEndpointSlicesV1beta1Informer := func() error {
				klog.Info("k3s start v1beta1.EndpointSlices informer")
				k3sEndpointSliceV1Beta1Informer := informerFactory.Discovery().V1beta1().EndpointSlices().Informer()

				// eventHandler
				k3sEndpointSliceV1Beta1Informer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						k3sEpsV1beta1 := obj.(*discoveryv1beta1.EndpointSlice)
						superedgeEpsV1beta1 := k3sEpsV1beta1.DeepCopy()
						superedgeEpsV1beta1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Name)
						superedgeEpsV1beta1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Labels["kubernetes.io/service-name"])
						s.endpointSliceV1Beta1WatchBroadcaster.ActionOrDrop(watch.Added, superedgeEpsV1beta1)
					},
					UpdateFunc: func(oldObj, newObj interface{}) {
						k3sEpsV1beta1 := newObj.(*discoveryv1beta1.EndpointSlice)
						superedgeEpsV1beta1 := k3sEpsV1beta1.DeepCopy()
						superedgeEpsV1beta1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Name)
						superedgeEpsV1beta1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Labels["kubernetes.io/service-name"])
						s.endpointSliceV1Beta1WatchBroadcaster.ActionOrDrop(watch.Modified, superedgeEpsV1beta1)
					},
					DeleteFunc: func(obj interface{}) {
						k3sEpsV1beta1 := obj.(*discoveryv1beta1.EndpointSlice)
						superedgeEpsV1beta1 := k3sEpsV1beta1.DeepCopy()
						superedgeEpsV1beta1.Name = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Name)
						superedgeEpsV1beta1.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", superedgeEpsV1beta1.Labels["kubernetes.io/service-name"])
						s.endpointSliceV1Beta1WatchBroadcaster.ActionOrDrop(watch.Deleted, superedgeEpsV1beta1)
					},
				}, resyncPeriod)
				// errorEventHandler
				k3sEndpointSliceV1Beta1Informer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
					for _, v := range k3sEndpointSliceV1Beta1Informer.GetStore().List() {
						k3sEndpointSliceV1Beta1Informer.GetStore().Delete(v)

						eps := v.(*discoveryv1beta1.EndpointSlice)
						k3sEps := eps.DeepCopy()
						k3sEps.Labels["kubernetes.io/service-name"] = fmt.Sprintf("k3s-%s", eps.Labels["kubernetes.io/service-name"])
						k3sEps.Name = fmt.Sprintf("k3s-%s", eps.Name)
						s.endpointSliceV1Beta1WatchBroadcaster.ActionOrDrop(watch.Deleted, k3sEps)
					}
				})
				go k3sEndpointSliceV1Beta1Informer.Run(k3sStop)
				if !cache.WaitForNamedCacheSync("k3s", k3sStop, k3sEndpointSliceV1Beta1Informer.HasSynced) {
					return fmt.Errorf("can't sync endpointslice informers")
				}
				s.k3sEndpointSliceV1Beta1Informer = k3sEndpointSliceV1Beta1Informer
				return nil
			}

			k3sRestMapperRes, err := restmapper.GetAPIGroupResources(k3sClient.Discovery())
			if err != nil {
				return err
			}
			k3sRestMapper := restmapper.NewDiscoveryRESTMapper(k3sRestMapperRes)
			_, err = k3sRestMapper.RESTMapping(apischema.GroupKind{
				Group: discoveryv1.SchemeGroupVersion.Group,
				Kind:  "EndpointSlice",
			}, discoveryv1.SchemeGroupVersion.Version)
			if err == nil {
				klog.Info("k3s mapper v1.EndpointSlices")
				err = setupEndpointSlicesV1Informer()
				if err != nil {
					klog.Error(err)
					return err
				}
			} else {
				if _, ok := err.(*meta.NoKindMatchError); ok {
					_, err = k3sRestMapper.RESTMapping(apischema.GroupKind{
						Group: discoveryv1beta1.SchemeGroupVersion.Group,
						Kind:  "EndpointSlice",
					}, discoveryv1beta1.SchemeGroupVersion.Version)
					if err == nil {
						klog.Info("mapper v1beta1.EndpointSlices")
						err = setupEndpointSlicesV1beta1Informer()
						if err != nil {
							klog.Error(err)
							return err
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
		}
		return nil
	}

	var oldk3sNodeUinit *sitev1alpha2.NodeUnit
	var newk3sNodeUinit *sitev1alpha2.NodeUnit
	var k3sInfomerStop chan struct{}

	go func() {
		checkTime := 10 * time.Second
		for true {
			node, err := nodeLister.Get(s.cache.GetNodeName())
			if err != nil {
				klog.Errorf("Failed to get node %s, error:%v", s.cache.GetNodeName(), err)
				time.Sleep(checkTime)
				continue
			}
			for k, v := range node.Labels {
				if v == siteconstant.NodeUnitSuperedge {
					nu, err := crdClient.SiteV1alpha2().NodeUnits().Get(context.Background(), k, metav1.GetOptions{})
					if err != nil {
						klog.Errorf("Failed to get nodeUint %s, error: %v", k, err)
						continue
					}
					if nu.Spec.Type != "edge" {
						continue
					}
					if nu.Spec.AutonomyLevel == sitev1alpha2.AutonomyLevelL4 || nu.Spec.AutonomyLevel == sitev1alpha2.AutonomyLevelL5 {
						newk3sNodeUinit = nu
					} else {
						if k3sInfomerStop != nil && (oldk3sNodeUinit != nil || newk3sNodeUinit != nil) {
							klog.Infof("Start closing the informer")
							k3sInfomerStop <- struct{}{}
							klog.Infof("close informer end")
							oldk3sNodeUinit = nil
							newk3sNodeUinit = nil
						}
					}
				}
				if newk3sNodeUinit != nil {
					if oldk3sNodeUinit != nil {
						if newk3sNodeUinit.Name == oldk3sNodeUinit.Name && newk3sNodeUinit.Spec.AutonomyLevel == oldk3sNodeUinit.Spec.AutonomyLevel {
							continue
						}
					}
					if newk3sNodeUinit.Spec.AutonomyLevel == sitev1alpha2.AutonomyLevelL4 || newk3sNodeUinit.Spec.AutonomyLevel == sitev1alpha2.AutonomyLevelL5 {
						kubeconf, err := client.CoreV1().ConfigMaps(newk3sNodeUinit.Spec.UnitCredentialConfigMapRef.Namespace).Get(context.Background(), newk3sNodeUinit.Spec.UnitCredentialConfigMapRef.Name, metav1.GetOptions{})
						if err != nil {
							klog.Errorf("Failed to get k3s kubeconfig, error: %v", err)
							continue
						}
						serviceLabelSelector := labels.NewSelector()
						nuRequirement, err := labels.NewRequirement(newk3sNodeUinit.Name, selection.Equals, []string{siteconstant.NodeUnitSuperedge})
						if err != nil {
							klog.Error(err)
							continue
						}
						serviceLabelSelector = serviceLabelSelector.Add(*nuRequirement)
						kinsLabelSelector, err := labels.NewRequirement(unitcluster.KinsResourceLabelKey, selection.Equals, []string{"yes"})
						if err != nil {
							klog.Error(err)
							continue
						}
						serviceLabelSelector = serviceLabelSelector.Add(*kinsLabelSelector)
						services, err := listersv1.NewServiceLister(serviceInformer.GetIndexer()).List(serviceLabelSelector)
						if err != nil {
							klog.Error(err)
							continue
						}
						var k3sSvc *v1.Service
						for _, svc := range services {
							if svc.Name == fmt.Sprintf("%s-svc-kins", newk3sNodeUinit.Name) {
								k3sSvc = svc
								break
							}
						}
						if k3sSvc == nil {
							klog.Infof("Failed to get k3sService, nodeunit:%s", newk3sNodeUinit.Name)
							continue
						}
						if k3sInfomerStop == nil {
							k3sInfomerStop = make(chan struct{})
						} else {
							k3sInfomerStop <- struct{}{}
						}
						err = setupK3sInformer(kubeconf.Data["kubeconfig.conf"], k, k3sSvc, k3sInfomerStop)
						if err != nil {
							klog.Errorf("Failed to setupK3sInformer, error: %v", err)
						} else {
							klog.Infof("setupK3sInformer Successfully, nodeUnit:%s", newk3sNodeUinit.Name)
						}
						oldk3sNodeUinit = newk3sNodeUinit
					}

				}
			}
			time.Sleep(checkTime)
		}
	}()

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

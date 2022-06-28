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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/klog/v2"

	"github.com/munnerz/goautoneg"
	"github.com/superedge/superedge/cmd/lite-apiserver/app/options"
	"github.com/superedge/superedge/pkg/lite-apiserver/cache"
	"github.com/superedge/superedge/pkg/lite-apiserver/cert"
	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/proxy"
	muxserver "github.com/superedge/superedge/pkg/lite-apiserver/server/multiplex"
	"github.com/superedge/superedge/pkg/lite-apiserver/storage"
	"github.com/superedge/superedge/pkg/lite-apiserver/transport"

	"github.com/superedge/superedge/pkg/util"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclientwatch "k8s.io/client-go/rest/watch"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// LiteServer ...
type LiteServer struct {
	ServerConfig *config.LiteServerConfig
	stopCh       <-chan struct{}
}

// CreateServer ...
func CreateServer(serverOptions *options.ServerRunOptions, stopCh <-chan struct{}) (*LiteServer, error) {

	config, err := serverOptions.ApplyTo()
	if err != nil {
		return nil, err
	}
	return &LiteServer{
		ServerConfig: config,
		stopCh:       stopCh,
	}, nil
}

// Run ...
func (s *LiteServer) Run() error {

	certChannel := make(chan string, 10)
	transportChannel := make(chan string, 10)
	stopCh := make(<-chan struct{})
	// init cert manager
	certManager := cert.NewCertManager(s.ServerConfig, certChannel)
	err := certManager.Init()
	if err != nil {
		klog.Errorf("Init certManager error: %v", err)
		return err
	}
	certManager.Start()

	// init transport manager
	transportManager := transport.NewTransportManager(s.ServerConfig, certManager, certChannel, transportChannel)
	err = transportManager.Init()
	if err != nil {
		klog.Errorf("Init transportManager error: %v", err)
		return err
	}
	transportManager.Start()

	// init storage
	storage := storage.CreateStorage(s.ServerConfig)
	// init cache manager
	cacheManager := cache.NewCacheManager(storage)

	edgeServerHandler, err := proxy.NewEdgeServerHandler(s.ServerConfig, transportManager, cacheManager, transportChannel)
	if err != nil {
		klog.Errorf("Create edgeServerHandler error: %v", err)
		return err
	}

	// Create server mux async
	go wait.PollUntil(time.Second*3, func() (bool, error) {

		// first check kubelet tls bootstrap finish
		if _, err := os.Stat(s.ServerConfig.TLSConfig[0].CertPath); err != nil {
			klog.V(4).Infof("Stat kubelet cert file %s error %v", s.ServerConfig.TLSConfig[0].CertPath, err)
			return false, err
		}

		// create client and informer factory
		restConfig, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*clientcmdapi.Config, error) {
			conf := s.generateKubeConfiguration()
			return conf, nil
		})
		if err != nil {
			klog.Errorf("clientcmd.BuildConfigFromKubeconfigGetter error: %v", err)
			return false, err
		}
		// get kubelet cert common name for transport
		var kubeletCertCommonName string
		for cn := range certManager.GetCertMap() {
			if cn != "" {
				kubeletCertCommonName = cn
				klog.V(5).Infof("kubelet cert CommonName %s", kubeletCertCommonName)
				break
			}
		}
		// replace restConfig transport
		restConfig.Wrap(func(rt http.RoundTripper) http.RoundTripper {
			// use transportManager default transport, it can reload cert
			return transportManager.GetTransport(kubeletCertCommonName)
		})
		restConfig.UserAgent = "lite-apiserver/mux"
		kubeClient := kubernetes.NewForConfigOrDie(restConfig)
		informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0)

		for _, url := range s.ServerConfig.URLMultiplexCache {
			mux, err := muxserver.CreateMux(url, "", informerFactory)
			if err != nil {
				klog.Errorf("Create url %s Mux error: %v", url, err)
				return false, err
			}
			muxserver.RegisterMux(url, mux)
		}
		if len(s.ServerConfig.URLMultiplexCache) > 0 {
			informerFactory.Start(stopCh)
			for t, hasSynced := range informerFactory.WaitForCacheSync(stopCh) {
				if !hasSynced {
					klog.Errorf("Sync informer %s cache failed", t.Name())
					return false, err
				}
			}
		}
		return true, nil
	}, make(<-chan struct{}))

	mux := http.NewServeMux()
	mux.Handle("/", edgeServerHandler)
	mux.HandleFunc("/debug/flags/v", util.UpdateLogLevel)
	// register for pprof
	if s.ServerConfig.Profiling {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	handler := http.Handler(mux)
	// check if url need cache multiplex
	handler = s.interceptCacheMultiplex(handler)
	handler = logger(handler)

	caCrt, err := ioutil.ReadFile(s.ServerConfig.CAFile)
	if err != nil {
		klog.Errorf("Read ca file err: %v", err)
		return err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCrt)

	for _, addr := range s.ServerConfig.ListenAddress {
		go func(addr string) {
			ser := &http.Server{
				Addr:    fmt.Sprintf("%s:%d", addr, s.ServerConfig.Port),
				Handler: handler,
				TLSConfig: &tls.Config{
					ClientCAs:  pool,
					ClientAuth: tls.VerifyClientCertIfGiven,
				},
			}

			klog.Infof("Listen on %s", ser.Addr)
			klog.Fatal(ser.ListenAndServeTLS(s.ServerConfig.CertFile, s.ServerConfig.KeyFile))
		}(addr)

	}

	<-s.stopCh
	klog.Info("Received a program exit signal")
	return nil
}

func logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("method %s, request: %s", r.Method, r.URL.String())
		handler.ServeHTTP(w, r)
	})
}

func (s *LiteServer) interceptCacheMultiplex(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only cache for GET method
		if r.Method != http.MethodGet {
			handler.ServeHTTP(w, r)
			return
		}
		// get mux by url
		mux, err := muxserver.GetMux(r.URL.Path)
		if err != nil {
			handler.ServeHTTP(w, r)
			return
		}
		klog.V(4).Infof("Multiplex Cache URL %s use %s Mux, will return data from lite-apiserver, URL.Path is %s, User-Agent is %s",
			mux.Name(), r.URL.String(), r.URL.Path, r.UserAgent())
		queries := r.URL.Query()
		acceptType := r.Header.Get("Accept")
		info, found := parseAccept(acceptType, scheme.Codecs.SupportedMediaTypes())
		if !found {
			klog.Errorf("can't find %s serializer", acceptType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// get label from param
		labelStr := queries.Get("labelSelector")
		ls, err := labels.Parse(labelStr)
		if err != nil {
			klog.Errorf("invalid labelSelector parameter %s", labelStr)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		rv := queries.Get("resourceVersion")
		var watchBookmarks bool
		wbStr := queries.Get("allowWatchBookmarks")
		if wbStr != "" {
			watchBookmarks, err = strconv.ParseBool(wbStr)
			if err != nil {
				klog.Errorf("invalid allowWatchBookmarks parameter %s", wbStr)
			}
		}
		encoder := scheme.Codecs.EncoderForVersion(info.Serializer, v1.SchemeGroupVersion)
		// list request
		if queries.Get("watch") == "" {
			w.Header().Set("Content-Type", info.MediaType)
			// list obj
			var objList runtime.Object
			switch r.URL.Path {
			// k8s list will return xxxList object instead []xxx, so we must know query path
			case muxserver.EndpointCacheURL:
				var eps []*v1.Endpoints
				err := mux.ListObjects(ls, func(m interface{}) {
					eps = append(eps, m.(*v1.Endpoints))
				})
				if err != nil {
					klog.Errorf("failed get resource list %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				epsItems := make([]v1.Endpoints, 0, len(eps))
				for _, ep := range eps {
					epsItems = append(epsItems, *ep)
				}
				// TODO build objectList and deal with fieldselector
				objList = &v1.EndpointsList{
					Items: epsItems,
				}

			case muxserver.NodeCacheURL:
				var nodes []*v1.Node
				err := mux.ListObjects(ls, func(m interface{}) {
					nodes = append(nodes, m.(*v1.Node))
				})
				if err != nil {
					klog.Errorf("failed get resource list %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				nodeItems := make([]v1.Node, 0, len(nodes))
				for _, n := range nodes {
					nodeItems = append(nodeItems, *n)
				}
				// TODO build objectList and deal with fieldselector
				objList = &v1.NodeList{
					Items: nodeItems,
				}
			case muxserver.ServiceCacheURL:
				var svcs []*v1.Service
				err := mux.ListObjects(ls, func(m interface{}) {
					svcs = append(svcs, m.(*v1.Service))
				})
				if err != nil {
					klog.Errorf("failed get resource list %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				svcItems := make([]v1.Service, 0, len(svcs))
				for _, s := range svcs {
					svcItems = append(svcItems, *s)
				}
				// TODO build objectList and deal with fieldselector
				objList = &v1.ServiceList{
					Items: svcItems,
				}

			}

			err := encoder.Encode(objList, w)
			if err != nil {
				klog.Errorf("can't marshal resource list, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		}

		// cacheMux.Watch label and field selector
		// watch request
		timeoutSecondsStr := r.URL.Query().Get("timeoutSeconds")
		timeout := time.Minute
		if timeoutSecondsStr != "" {
			timeout, _ = time.ParseDuration(fmt.Sprintf("%ss", timeoutSecondsStr))
		}

		timer := time.NewTimer(timeout)
		defer timer.Stop()

		flusher, ok := w.(http.Flusher)
		if !ok {
			klog.Errorf("unable to start watch - can't get http.Flusher: %#v", w)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		muxWatcher, err := mux.Watch(watchBookmarks, rv)
		if err != nil {
			klog.Errorf("unable to start mux watch: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer muxWatcher.Stop()
		wacherChan := muxWatcher.ResultChan()

		e := restclientwatch.NewEncoder(
			streaming.NewEncoder(info.StreamSerializer.Framer.NewFrameWriter(w),
				scheme.Codecs.EncoderForVersion(info.StreamSerializer, v1.SchemeGroupVersion)),
			encoder)
		if info.MediaType == runtime.ContentTypeProtobuf {
			w.Header().Set("Content-Type", runtime.ContentTypeProtobuf+";stream=watch")
		} else {
			w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		}
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-timer.C:
				return
			case evt := <-wacherChan:
				klog.V(5).Infof("Send watch event type: %+#v, event object %+#v", evt.Type, evt.Object)
				err := e.Encode(&evt)
				if err != nil {
					klog.Errorf("can't encode watch event, %v", err)
					return
				}

				if len(wacherChan) == 0 {
					flusher.Flush()
				}
			}
		}
	})

}
func parseAccept(header string, accepted []runtime.SerializerInfo) (runtime.SerializerInfo, bool) {
	if len(header) == 0 && len(accepted) > 0 {
		return accepted[0], true
	}

	clauses := goautoneg.ParseAccept(header)
	for i := range clauses {
		clause := &clauses[i]
		for i := range accepted {
			accepts := &accepted[i]
			switch {
			case clause.Type == accepts.MediaTypeType && clause.SubType == accepts.MediaTypeSubType,
				clause.Type == accepts.MediaTypeType && clause.SubType == "*",
				clause.Type == "*" && clause.SubType == "*":
				return *accepts, true
			}
		}
	}

	return runtime.SerializerInfo{}, false
}

func (s *LiteServer) generateKubeConfiguration() *clientcmdapi.Config {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:               fmt.Sprintf("https://%s:%d", s.ServerConfig.KubeApiserverUrl, s.ServerConfig.KubeApiserverPort),
		CertificateAuthority: s.ServerConfig.CAFile,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:  "default-cluster",
		AuthInfo: "default-user",
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		ClientKey:         s.ServerConfig.TLSConfig[0].KeyPath,
		ClientCertificate: s.ServerConfig.TLSConfig[0].CertPath,
	}

	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	return &clientConfig
}

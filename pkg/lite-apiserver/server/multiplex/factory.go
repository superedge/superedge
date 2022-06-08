package multiplex

import (
	"fmt"

	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
)

var muxFactories = make(map[string]CacheMuxFactory)
var muxRegistries = make(map[string]CacheMux)

type CacheMuxFactory interface {
	Create(hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error)
}

func RegisterMuxFactory(url string, f CacheMuxFactory) {
	if f == nil {
		panic("Must provide valid CacheMuxFactory")
	}
	_, registered := muxFactories[url]
	if registered {
		panic(fmt.Sprintf("CacheMuxFactory url %s already registered", url))
	}

	muxFactories[url] = f
}

func CreateMux(url, hostname string, informerFactory informers.SharedInformerFactory) (CacheMux, error) {
	muxFactory, ok := muxFactories[url]
	if !ok {
		return nil, fmt.Errorf("Unsupport Multiplex URL %s", url)
	}
	return muxFactory.Create(hostname, informerFactory)
}
func GetMuxFactory(url string) (CacheMuxFactory, error) {
	muxFactory, ok := muxFactories[url]
	if !ok {
		return nil, fmt.Errorf("Unsupport Multiplex URL %s", url)
	}
	return muxFactory, nil
}

func GetMux(url string) (CacheMux, error) {
	mux, ok := muxRegistries[url]
	if !ok {
		return nil, fmt.Errorf("Unregister Multiplex URL")
	}
	return mux, nil
}

func RegisterMux(url string, mux CacheMux) {
	klog.V(4).Infof("Register URL %s for mux cache", url)
	muxRegistries[url] = mux
}

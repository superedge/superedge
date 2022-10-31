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
	"fmt"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type upstream struct {
	config       *rest.Config
	backendProxy *httputil.ReverseProxy
	nodech       <-chan watch.Event
}

func Newupstream(restConfig *rest.Config) (*upstream, error) {

	tr, err := rest.TransportFor(restConfig)
	if err != nil {
		klog.Errorf("")

	}

	up := &upstream{
		config: restConfig,
	}

	reverseProxy := &httputil.ReverseProxy{
		Director:  up.makeDirector,
		Transport: tr,
	}
	up.backendProxy = reverseProxy
	return up, nil
}

func (up *upstream) makeDirector(req *http.Request) {
	url, err := url.Parse(up.config.Host)
	if err != nil {
		klog.Errorf("Failed to get apiserver url, error: %v", err)
		backendUrl := os.Getenv("KUBERNETES_SERVICE_HOST")
		backendPort := os.Getenv("KUBERNETES_SERVICE_PORT")
		req.URL.Host = fmt.Sprintf("%s:%s", backendUrl, backendPort)
	} else {
		req.URL.Host = url.Host
	}
	req.URL.Scheme = "https"
}

func (up *upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	up.backendProxy.ServeHTTP(w, r)
}

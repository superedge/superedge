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

package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/superedge/superedge/pkg/edge-health-admission/admission"
	"github.com/superedge/superedge/pkg/edge-health-admission/config"
	"github.com/superedge/superedge/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	Certconfig                 config.Config
	admissionControlListenAddr string
)

func init() {
	flag.StringVar(&admissionControlListenAddr, "adminssion-control-listen-addr", ":8443", "")
	flag.BoolVar(&config.NodeAlwaysReachable, "always-reachable", false, "set true addmision will forbidden apiserver set Unreachable taint, when node can't access apiserver")
}

func main() {
	klog.InitFlags(nil)

	Certconfig.AddFlags()

	flag.Parse()

	klog.Infof("Versions: %#v\n", version.Get())

	klog.V(4).Infof("master url is %s", Certconfig.MasterUrl)
	config.Kubeclient = generateClientset(Certconfig.MasterUrl, Certconfig.KubeconfigPath)

	http.HandleFunc("/node-taint", admission.NodeTaint)
	http.HandleFunc("/endpoint", admission.EndPoint)
	server := &http.Server{
		Addr: admissionControlListenAddr,
	}
	err := server.ListenAndServeTLS(Certconfig.CertFile, Certconfig.KeyFile)
	if err != nil {
		time.Sleep(time.Duration(10) * time.Second)
		klog.Errorf("ListenAndServeTLS err: %s", err.Error())
	}
}

func generateClientset(masterUrl, kubeconfigPath string) *kubernetes.Clientset {
	var err error
	kubeconfig, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		klog.Fatalf("Init: Error building kubeconfig: %s", err.Error())
	}
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Init: Error building clientset: %s", err.Error())
	}
	return clientset
}

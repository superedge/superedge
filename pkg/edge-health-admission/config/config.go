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

package config

import (
	"github.com/superedge/superedge/cmd/edge-health-admission/app/options"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"time"
)

// EdgeHealthAdmissionConfig contains both cert and key of admission webhook
type EdgeHealthAdmissionConfig struct {
	CertFile     string
	KeyFile      string
	Addr         string
	NodeInformer corev1.NodeInformer
}

func NewEdgeHealthAdmissionConfig(o options.CompletedOptions) (*EdgeHealthAdmissionConfig, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(o.Master, o.Kubeconfig)
	if err != nil {
		klog.Errorf("Building kubeconfig error %s", err.Error())
		return nil, err
	}
	kubeconfig.QPS = o.QPS
	kubeconfig.Burst = o.Burst
	kubeclient, err := clientset.NewForConfig(kubeconfig)
	if err != nil {
		klog.Errorf("Building kubeclient error %s", err.Error())
		return nil, err
	}
	k8sFactory := informers.NewSharedInformerFactory(kubeclient, time.Second*time.Duration(o.SyncPeriod))

	return &EdgeHealthAdmissionConfig{
		CertFile:     o.CertFile,
		KeyFile:      o.KeyFile,
		Addr:         o.Addr,
		NodeInformer: k8sFactory.Core().V1().Nodes(),
	}, nil
}

func (c *EdgeHealthAdmissionConfig) Run(stop <-chan struct{}) {
	go c.NodeInformer.Informer().Run(stop)
}

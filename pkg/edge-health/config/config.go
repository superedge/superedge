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
	"context"
	"github.com/superedge/superedge/cmd/edge-health/app/options"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"time"
)

type EdgeHealthConfig struct {
	Kubeclient        clientset.Interface
	ConfigMapInformer corev1.ConfigMapInformer
	NodeInformer      corev1.NodeInformer
	Check             EdgeHealthCheck
	Commun            EdgeHealthCommun
	Node              EdgeHealthNode
	Vote              EdgeHealthVote
}

type EdgeHealthCheck struct {
	HealthCheckPeriod    int
	HealthCheckScoreLine float64
}

type EdgeHealthCommun struct {
	CommunPeriod     int
	CommunTimeout    int
	CommunRetries    int
	CommunServerPort int
}

type EdgeHealthNode struct {
	HostName string // Node name
	LocalIp  string // Node IP
	Worker   int    // Reserved for future use
}

type EdgeHealthVote struct {
	VotePeriod  int
	VoteTimeout int // Vote will be timeout after VoteTimeout
}

func NewEdgeHealthConfig(o options.CompletedOptions) (*EdgeHealthConfig, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(o.Node.MasterUrl, o.Node.KubeconfigPath)
	if err != nil {
		klog.Errorf("Building kubeconfig error %s", err.Error())
		return nil, err
	}
	kubeconfig.QPS = o.Node.QPS
	kubeconfig.Burst = o.Node.Burst
	kubeclient, err := clientset.NewForConfig(kubeconfig)
	if err != nil {
		klog.Errorf("Building kubeclient error %s", err.Error())
		return nil, err
	}
	k8sFactory := informers.NewSharedInformerFactory(kubeclient, time.Second*time.Duration(o.Node.SyncPeriod))

	var localIp string
	if host, err := kubeclient.CoreV1().Nodes().Get(context.TODO(), o.Node.HostName, metav1.GetOptions{}); err != nil {
		return nil, err
	} else {
		for _, v := range host.Status.Addresses {
			if v.Type == v1.NodeInternalIP {
				localIp = v.Address
				klog.V(2).Infof("Host ip is %s", localIp)
			}
		}
	}

	return &EdgeHealthConfig{
		Kubeclient:        kubeclient,
		ConfigMapInformer: k8sFactory.Core().V1().ConfigMaps(),
		NodeInformer:      k8sFactory.Core().V1().Nodes(),
		Check: EdgeHealthCheck{
			HealthCheckPeriod:    o.Check.CheckPeriod,
			HealthCheckScoreLine: o.Check.CheckScoreLine,
		},
		Commun: EdgeHealthCommun{
			CommunPeriod:     o.Commun.CommunPeriod,
			CommunTimeout:    o.Commun.CommunTimeout,
			CommunRetries:    o.Commun.CommunRetries,
			CommunServerPort: o.Commun.CommunServerPort,
		},
		Node: EdgeHealthNode{
			HostName: o.Node.HostName,
			LocalIp:  localIp,
			Worker:   o.Node.Worker,
		},
		Vote: EdgeHealthVote{
			VotePeriod:  o.Vote.VotePeriod,
			VoteTimeout: o.Vote.VoteTimeout,
		},
	}, nil
}

func (c *EdgeHealthConfig) Run(stop <-chan struct{}) {
	go c.ConfigMapInformer.Informer().Run(stop)
	go c.NodeInformer.Informer().Run(stop)
}

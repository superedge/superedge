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

package connect

import (
	"bufio"
	"bytes"
	cctx "context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"net"
	"os"
	"strings"
	"superedge/pkg/tunnel/conf"
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/util"
	"time"
)

var coreDns *CoreDns

type CoreDns struct {
	PodIp     string
	Namespace string
	ClientSet *kubernetes.Clientset
	Update    chan struct{}
}

func InitDNS() error {
	coreDns = &CoreDns{
		Update: make(chan struct{}),
	}
	coreDns.PodIp = os.Getenv(util.POD_IP_ENV)
	klog.Infof("endpoint of the proxycloud pod = %s ", coreDns.PodIp)
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("client-go get inclusterconfig  fail err = %v", err)
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("get client fail err = %v", err)
		return err
	}
	coreDns.ClientSet = clientset
	coreDns.Namespace = os.Getenv(util.POD_NAMESPACE_ENV)
	return nil
}

func (dns *CoreDns) checkHosts() error {
	nodes, flag := parseHosts()
	if !flag {
		return nil
	}
	var hostsBuffer bytes.Buffer
	for k, v := range nodes {
		hostsBuffer.WriteString(v)
		hostsBuffer.WriteString("    ")
		hostsBuffer.WriteString(k)
		hostsBuffer.WriteString("\n")
	}
	cm, err := dns.ClientSet.CoreV1().ConfigMaps(dns.Namespace).Get(cctx.TODO(), conf.TunnelConf.TunnlMode.Cloud.Stream.Dns.Configmap, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("get configmap fail err = %v", err)
		return err
	}
	if hostsBuffer.Len() != 0 {
		cm.Data[util.COREFILE_HOSTS_FILE] = hostsBuffer.String()
	} else {
		cm.Data[util.COREFILE_HOSTS_FILE] = ""
	}
	_, err = dns.ClientSet.CoreV1().ConfigMaps(dns.Namespace).Update(cctx.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("update configmap fail err = %v", err)
		return err
	}
	klog.Infof("update configmap success!")
	return nil
}

func SynCorefile() {
	for {
		klog.V(8).Infof("connected node total = %d nodes = %v", len(context.GetContext().GetNodes()), context.GetContext().GetNodes())
		err := coreDns.checkHosts()
		if err != nil {
			klog.Errorf("failed to synchronize hosts periodically err = %v", err)
		}
		time.Sleep(60 * time.Second)
	}
}

func parseHosts() (map[string]string, bool) {
	file, err := os.Open(conf.TunnelConf.TunnlMode.Cloud.Stream.Dns.Hosts)
	if err != nil {
		klog.Errorf("load hosts fail! err = %v", err)
		return nil, false
	}
	scanner := bufio.NewScanner(file)
	eps, err := coreDns.ClientSet.CoreV1().Endpoints(coreDns.Namespace).Get(cctx.Background(), conf.TunnelConf.TunnlMode.Cloud.Stream.Dns.Service, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("failed to get %s endpoint ip err = %v", conf.TunnelConf.TunnlMode.Cloud.Stream.Dns.Service, err)
		return nil, false
	}
	existCount := 0
	disconnectCount := 0
	nodes := make(map[string]string)
	update := false
	for scanner.Scan() {
		line := scanner.Bytes()
		f := bytes.Fields(line)
		if len(f) < 2 {
			update = true
			continue
		}
		addr := parseIP(string(f[0]))
		if addr == nil {
			update = true
			continue
		}
		if addr.String() == coreDns.PodIp {
			if !update {
				if context.GetContext().NodeIsExist(string(f[1])) {
					existCount += 1
				} else {
					disconnectCount += 1
				}
			}
			continue
		} else {
			orphan := true
			for _, ipv := range eps.Subsets[0].Addresses {
				if addr.String() == ipv.IP {
					if context.GetContext().NodeIsExist(string(f[1])) {
						update = true
					} else {
						nodes[string(f[1])] = addr.String()
					}
					orphan = false
					break
				}
			}
			if orphan {
				update = true
			}
		}

	}
	file.Close()
	if update {
		for _, v := range context.GetContext().GetNodes() {
			nodes[v] = coreDns.PodIp
		}
	} else {
		if existCount != len(context.GetContext().GetNodes()) || disconnectCount != 0 {
			for _, v := range context.GetContext().GetNodes() {
				nodes[v] = coreDns.PodIp
			}
			update = true
		}
	}
	return nodes, update
}

func parseIP(addr string) net.IP {
	if i := strings.Index(addr, "%"); i >= 0 {
		addr = addr[0:i]
	}
	return net.ParseIP(addr)
}

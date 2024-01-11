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
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"github.com/superedge/superedge/pkg/tunnel/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var register *RegisterNode

type RegisterNode struct {
	ClientSet *kubernetes.Clientset
	Update    chan struct{}
}

func InitRegister() error {
	register = &RegisterNode{
		Update: make(chan struct{}),
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.ErrorS(err, "failed to get InClusterConfig")
		return err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.ErrorS(err, "failed to get  clientSet")
		return err
	}
	register.ClientSet = clientSet
	return nil
}

func (registerNode *RegisterNode) syncPodIP() error {
	file, err := os.Open(util.HostsPath)
	if err != nil {
		klog.ErrorS(err, "failed to load hosts")
		return err
	}
	defer file.Close()

	arrays := hosts2Array(file)
	_, update := filterPodIp(arrays)
	if !update {
		return nil
	}

	err = wait.Poll(2*time.Second, 10*time.Second, func() (done bool, err error) {
		cm, err := registerNode.ClientSet.CoreV1().ConfigMaps(os.Getenv(util.POD_NAMESPACE_ENV)).Get(cctx.TODO(), util.HostsConfig, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("get configmap fail err = %v", err)
			return false, err
		}
		arrays := hosts2Array(strings.NewReader(cm.Data[util.COREFILE_HOSTS_FILE]))
		hosts, flag := filterPodIp(arrays)
		if flag {
			cm.Data[util.COREFILE_HOSTS_FILE] = hosts
			_, err = registerNode.ClientSet.CoreV1().ConfigMaps(os.Getenv(util.POD_NAMESPACE_ENV)).Update(cctx.TODO(), cm, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("update configmap fail err = %v", err)
				return false, err
			}
		}
		return true, nil
	})
	return err
}

func (registerNode *RegisterNode) syncEndpoints() error {
	file, err := os.Open(util.HostsPath)
	if err != nil {
		klog.Errorf("load hosts fail! err = %v", err)
		return err
	}
	defer file.Close()

	arrays := hosts2Array(file)
	_, update := filterEndpoint(arrays)
	if !update {
		return nil
	}
	err = wait.Poll(5*time.Second, 30*time.Second, func() (done bool, err error) {
		cm, err := registerNode.ClientSet.CoreV1().ConfigMaps(os.Getenv(util.POD_NAMESPACE_ENV)).Get(cctx.TODO(), util.HostsConfig, metav1.GetOptions{})
		if err != nil {
			klog.ErrorS(err, "failed to get configmap", "configmap", util.HostsConfig)
			return false, err
		}
		arrays := hosts2Array(strings.NewReader(cm.Data[util.COREFILE_HOSTS_FILE]))
		hosts, flag := filterEndpoint(arrays)
		if flag {
			cm.Data[util.COREFILE_HOSTS_FILE] = hosts
			_, err = registerNode.ClientSet.CoreV1().ConfigMaps(os.Getenv(util.POD_NAMESPACE_ENV)).Update(cctx.TODO(), cm, metav1.UpdateOptions{})
			if err != nil {
				klog.Errorf("update configmap fail err = %v", err)
				return false, err
			}
		}
		return true, nil
	})
	return err
}

func SyncPodIP() {
	for {
		klog.V(3).InfoS("connected node", "number", len(tunnelcontext.GetContext().GetNodes()), "nodes", tunnelcontext.GetContext().GetNodes())
		err := register.syncPodIP()
		if err != nil {
			klog.ErrorS(err, "failed to synchronize hosts periodically")
		}
		err = loadCacheFromLocalFile()
		if err != nil {
			klog.ErrorS(err, "failed to load route from files")
		}
		time.Sleep(60 * time.Second)
	}
}

func SyncEndPoints() {
	for {
		time.Sleep(10 * time.Minute)
		klog.V(3).InfoS("connected node", "number", len(tunnelcontext.GetContext().GetNodes()), "nodes", tunnelcontext.GetContext().GetNodes())
		err := register.syncEndpoints()
		if err != nil {
			klog.Errorf("failed to synchronize endpoints periodically err = %v", err)
		}
	}
}

// Read the file by line, split each line of data read according to the space,
// and the variable after splitting is a byte array
// e: 127.0.0.1    localhsot
// hostsArray = [[[49 50 55 46 48 46 48 46 49] [108 111 99 97 108 104 115 111 116]]]
func hosts2Array(fileread io.Reader) [][][]byte {
	scanner := bufio.NewScanner(fileread)
	hostsArray := [][][]byte{}
	for scanner.Scan() {
		// copy byte slice before append to hostsArray
		f := bytes.Fields([]byte(scanner.Text()))
		if len(f) < 2 {
			hostsArray = append(hostsArray, f)
			continue
		}
		addr := parseIP(string(f[0]))
		if addr == nil {
			continue
		}
		hostsArray = append(hostsArray, f)
	}
	return hostsArray
}

func filterEndpoint(hostsArray [][][]byte) (string, bool) {
	var eps *v1.Endpoints
	var err error
	hostsBuffer := &bytes.Buffer{}
	if register != nil {
		eps, err = register.ClientSet.CoreV1().Endpoints(os.Getenv(util.POD_NAMESPACE_ENV)).Get(cctx.Background(), conf.TunnelConf.TunnlMode.Cloud.Stream.Register.Service, metav1.GetOptions{})
		if err != nil {
			klog.ErrorS(err, "failed to get endpoints", "endpoints", conf.TunnelConf.TunnlMode.Cloud.Stream.Register.Service)
			return "", false
		}
	}
	update := false
	for _, v := range hostsArray {
		if len(v) < 2 {
			continue
		}
		addr := parseIP(string(v[0]))
		if addr == nil {
			continue
		}
		flag := true
		for _, ep := range eps.Subsets[0].Addresses {
			if ep.IP == addr.String() {
				writeLine(v, hostsBuffer)
				flag = false
			}
		}
		if flag && !update {
			update = true
		}
	}
	if update {
		return hostsBuffer.String(), update
	}

	return "", update
}

func filterPodIp(hostsArray [][][]byte) (string, bool) {
	podIp := os.Getenv(util.POD_IP_ENV)
	hostsBuffer := &bytes.Buffer{}
	update := false
	nodes := 0

	for _, v := range hostsArray {
		if len(v) < 2 {
			continue
		}
		addr := parseIP(string(v[0]))
		if addr == nil {
			continue
		}
		if addr.String() == podIp {
			if tunnelcontext.GetContext().NodeIsExist(string(v[1])) {
				nodes++
			} else {
				update = true
			}
		} else {
			if tunnelcontext.GetContext().NodeIsExist(string(v[1])) {
				continue
			}
			writeLine(v, hostsBuffer)
		}
	}

	edgeNodes := tunnelcontext.GetContext().GetNodes()
	if update == false && nodes != len(edgeNodes) {
		update = true
	}

	if update {
		for _, v := range edgeNodes {
			hostsBuffer.WriteString(podIp)
			hostsBuffer.WriteString("    ")
			hostsBuffer.WriteString(v)
			hostsBuffer.WriteString("\n")
		}
		return hostsBuffer.String(), update
	} else {
		return "", update
	}
}

func parseIP(addr string) net.IP {
	if i := strings.Index(addr, "%"); i >= 0 {
		addr = addr[0:i]
	}
	return net.ParseIP(addr)
}

func writeLine(line [][]byte, buf *bytes.Buffer) {
	for k, v := range line {
		if k == len(line)-1 {
			buf.WriteString(string(v))
			buf.WriteString("\n")
		} else {
			buf.WriteString(string(v))
			buf.WriteString("    ")
		}
	}
}

func IsEndpointIp(addr string) bool {
	if register != nil {
		eps, err := register.ClientSet.CoreV1().Endpoints(os.Getenv(util.POD_NAMESPACE_ENV)).Get(cctx.Background(), conf.TunnelConf.TunnlMode.Cloud.Stream.Register.Service, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Failed to get SVC:%s endpoints, error: %v", conf.TunnelConf.TunnlMode.Cloud.Stream.Register.Service, err)
			return false
		}
		for _, ipv := range eps.Subsets[0].Addresses {
			if ipv.IP == addr {
				return true
			}
		}
	}
	return false
}

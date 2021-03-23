package edgecluster

import (
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
	"time"
)

const (
	annotationPrefix = "edge.tke.cloud.tencent.com"

	resolveIp                     = "resolveip"
	podCIDR                       = "pod-cidr"
	serviceCIDR                   = "service-cidr"
	apiserverExposeAddress        = "apiserver-expose-address"
	apiserverExposeAddressUrl     = "apiserver-expose-address-url"
	apiserverExposeAddressUrlPort = "apiserver-expose-address-url-port"
	apiserverExposeAddressType    = "apiserver-expose-address-type"
)

func (in *EdgeCluster) ClientSet() (*kubernetes.Clientset, error) {
	if in.Spec.Credential.Token == nil {
		return nil, errors.New("Get edge cluster token nil\n")
	}

	restConfig := &rest.Config{
		Host:        in.Address(AddressInternal).String(),
		BearerToken: *in.Spec.Credential.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		Timeout: 5 * time.Second,
	}
	return kubernetes.NewForConfig(restConfig)
}

func (in *EdgeCluster) RestConfig() *rest.Config {
	return &rest.Config{
		Host:        in.Address(AddressInternal).String(),
		BearerToken: *in.Spec.Credential.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		Timeout: 5 * time.Second,
	}
}

func (in *EdgeCluster) CenterAddrPort() (string, string) {
	var splitNResult []string
	edgeAccessCenter := in.Spec.EdgeAccessCenter
	if edgeAccessCenter.AccessType == AccessTypeGateway && len(edgeAccessCenter.AccessAddr) >= 1 {
		splitNResult = strings.SplitN(edgeAccessCenter.AccessAddr[0], ":", -1)
	}

	if len(splitNResult) == 2 {
		return splitNResult[0], splitNResult[1]
	}

	return splitNResult[0], ""
}

func (in *EdgeCluster) Address(addrType AddressType) Address {
	address := Address{}
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			address = Address(one)
			return address
		}
	}

	return address
}

func (in *EdgeCluster) GetMasterHosts(addrType AddressType) []string {
	var hosts []string
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			hosts = append(hosts, one.Host)
		}
	}
	return hosts
}

func (in *EdgeCluster) GetAdvertiseAddress(addrType AddressType) []ClusterAddress {
	var advertiseAddress []ClusterAddress
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			advertiseAddress = append(advertiseAddress, one)
		}
	}
	return advertiseAddress
}

func (in *EdgeCluster) AddAddress(addrType AddressType, host string, port int32) {
	addr := ClusterAddress{
		Type: addrType,
		Host: host,
		Port: port,
	}
	// skip same address
	for _, one := range in.Status.Addresses {
		if one == addr {
			return
		}
	}
	in.Status.Addresses = append(in.Status.Addresses, addr)
}

func (in *EdgeCluster) RemoveAddress(addrType AddressType) {
	var addrs []ClusterAddress
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			continue
		}
		addrs = append(addrs, one)
	}
	in.Status.Addresses = addrs
}

func (in *EdgeCluster) SetCondition(newCondition ClusterCondition) {
	var conditions []ClusterCondition
	exist := false
	for _, condition := range in.Status.Conditions {
		if condition.Type == newCondition.Type {
			exist = true
			if newCondition.Status != condition.Status {
				condition.Status = newCondition.Status
			}
			if newCondition.Message != condition.Message {
				condition.Message = newCondition.Message
			}
			if newCondition.Reason != condition.Reason {
				condition.Reason = newCondition.Reason
			}
			if !newCondition.LastProbeTime.IsZero() && newCondition.LastProbeTime != condition.LastProbeTime {
				condition.LastProbeTime = newCondition.LastProbeTime
			}
			if !newCondition.LastTransitionTime.IsZero() && newCondition.LastTransitionTime != condition.LastTransitionTime {
				condition.LastTransitionTime = newCondition.LastTransitionTime
			}
		}
		conditions = append(conditions, condition)
	}
	if !exist {
		if newCondition.LastProbeTime.IsZero() {
			newCondition.LastProbeTime = metav1.Now()
		}
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.Now()
		}
		conditions = append(conditions, newCondition)
	}
	in.Status.Conditions = conditions
}

type Address ClusterAddress

func (ca Address) String() string {
	return fmt.Sprintf("https://%s:%d", ca.Host, ca.Port)
}

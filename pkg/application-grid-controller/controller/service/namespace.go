package service

import (
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog/v2"
)

func (sgc *ServiceGridController) addNameSpace(obj interface{}) {
	ns := obj.(*corev1.Namespace)
	klog.V(4).Infof("Adding NameSpace %s", ns.Name)

	if _, ok := ns.Labels[common.FedManagedClustIdKey]; !ok {
		return
	}

	labelSelector := labels.NewSelector()
	gridRequirement, err := labels.NewRequirement(common.FedrationKey, selection.Equals, []string{"yes"})
	if err != nil {
		klog.V(4).Infof("Adding NameSpace %s error: gererate requirement err %v", ns.Name, err)
		return
	}
	labelSelector = labelSelector.Add(*gridRequirement)

	sgList, err := sgc.svcGridLister.List(labelSelector)
	if err != nil {
		klog.V(4).Infof("Adding NameSpace %s error: get sgList err %v", ns.Name, err)
		return
	}

	for _, sg := range sgList {
		sgc.enqueueServiceGrid(sg)
	}
}

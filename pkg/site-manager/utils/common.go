package utils

import (
	"context"

	sitev1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	crdClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	AllNodeUnit    = "unit-node-all"
	EdgeNodeUnit   = "unit-node-edge"
	CloudNodeUnit  = "unit-node-cloud"
	MasterNodeUnit = "unit-node-master"
)

func CreateDefaultUnit(crdClient *crdClientset.Clientset) error {
	// All Node Unit
	allNodeUnitSelector := &sitev1.Selector{
		MatchLabels: map[string]string{
			"kubernetes.io/os": "linux",
		},
	}
	allNodeUnit := &sitev1.NodeUnit{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "site.superedge.io/v1alpha2",
			Kind:       "NodeUnit",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: AllNodeUnit,
		},
		Spec: sitev1.NodeUnitSpec{
			Type:     sitev1.OtherNodeUnit,
			Selector: allNodeUnitSelector,
		},
	}

	if _, err := crdClient.SiteV1alpha2().NodeUnits().Create(context.TODO(), allNodeUnit, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		klog.Warningf("Create default %s unit error : %#v", AllNodeUnit, err)
	}

	return nil
}

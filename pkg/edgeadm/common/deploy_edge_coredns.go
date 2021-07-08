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
package common

import (
	"context"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	v1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	serviceGrpupClinet "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

// DeployEdgeCorednsAddon installs edge node CoreDNS addon to a Kubernetes cluster
func DeployEdgeCorednsAddon(kubeconfigFile string, manifestsDir string) error {
	client, err := kubeclient.GetClientSet(kubeconfigFile)
	if err != nil {
		return err
	}

	if err := EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	// Deploy edge-coredns config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgeSystem,
	}
	userEdgeCorednsConfig := filepath.Join(manifestsDir, manifests.APPEdgeCorednsConfig)
	edgeCorednsConfig := ReadYaml(userEdgeCorednsConfig, manifests.EdgeCorednsConfigYaml)
	// Waiting DeploymentGrid apply success
	err = kubeclient.CreateResourceWithFile(client, edgeCorednsConfig, option)
	if err != nil {
		klog.Errorf("Deploy edge-coredns config error: %v", err)
		return err
	}

	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return err
	}
	serviceGrpupClinet := serviceGrpupClinet.NewForConfigOrDie(restCfg)

	// Deploy edge-coredns deploymentGrid
	userCorednsDeploymentGrid := filepath.Join(manifestsDir, manifests.APPEdgeCorednsDeploymentGrid)
	edgeCorednsDeploymentGrid := ReadYaml(userCorednsDeploymentGrid, manifests.EdgeCorednsDeploymentGridYaml)
	data, err := kubeclient.ParseString(edgeCorednsDeploymentGrid, option)
	if err != nil {
		return err
	}

	obj := new(v1.DeploymentGrid)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return err
	}
	err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (bool, error) {
		_, err := serviceGrpupClinet.SuperedgeV1().DeploymentGrids(constant.NamespaceEdgeSystem).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Waiting deploy edge-coredns DeploymentGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Deploy %s success!", manifests.APPEdgeCorednsDeploymentGrid)

	// Deploy edge-coredns serviceGrid
	userCorednsServiceGrid := filepath.Join(manifestsDir, manifests.APPEdgeCorednsServiceGrid)
	edgeCorednsServiceGrid := ReadYaml(userCorednsServiceGrid, manifests.EdgeCorednsServiceGridYaml)
	data, err = kubeclient.ParseString(edgeCorednsServiceGrid, option)
	if err != nil {
		return err
	}

	serviceGrid := new(v1.ServiceGrid)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, serviceGrid); err != nil {
		return err
	}
	err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (bool, error) {
		_, err := serviceGrpupClinet.SuperedgeV1().ServiceGrids(constant.NamespaceEdgeSystem).Create(context.TODO(), serviceGrid, metav1.CreateOptions{})
		if err != nil {
			klog.V(2).Infof("Waiting deploy edge-coredns ServiceGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Deploy %s success!", manifests.APPEdgeCorednsServiceGrid)

	return err
}

// DeleteEdgeCoredns uninstalls edge node CoreDNS addon to a Kubernetes cluster
func DeleteEdgeCoredns(kubeconfigFile string, manifestsDir string) error {
	client, err := kubeclient.GetClientSet(kubeconfigFile)
	if err != nil {
		return err
	}

	if err := EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	// Delete edge-coredns
	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgeSystem,
	}
	userEdgeCorednsConfig := filepath.Join(manifestsDir, manifests.APPEdgeCorednsConfig)
	edgeCorednsConfig := ReadYaml(userEdgeCorednsConfig, manifests.EdgeCorednsConfigYaml)
	// Waiting DeploymentGrid apply success
	err = kubeclient.DeleteResourceWithFile(client, edgeCorednsConfig, option)
	if err != nil {
		klog.Errorf("Deploy edge-coredns config error: %v", err)
		return err
	}

	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return err
	}
	serviceGrpupClinet := serviceGrpupClinet.NewForConfigOrDie(restCfg)

	// Delete edge-coredns deploymentGrid
	userCorednsDeploymentGrid := filepath.Join(manifestsDir, manifests.APPEdgeCorednsDeploymentGrid)
	edgeCorednsDeploymentGrid := ReadYaml(userCorednsDeploymentGrid, manifests.EdgeCorednsDeploymentGridYaml)
	data, err := kubeclient.ParseString(edgeCorednsDeploymentGrid, option)
	if err != nil {
		return err
	}

	obj := new(v1.DeploymentGrid)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return err
	}
	err = wait.PollImmediate(time.Second, 5*time.Minute, func() (bool, error) {
		err := serviceGrpupClinet.SuperedgeV1().DeploymentGrids(constant.NamespaceEdgeSystem).Delete(context.TODO(), "", metav1.DeleteOptions{})
		if err != nil {
			klog.Warningf("Waiting deploy edge-coredns DeploymentGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Delete %s success!", manifests.APPEdgeCorednsDeploymentGrid)

	// Delete edge-coredns serviceGrid
	userCorednsServiceGrid := filepath.Join(manifestsDir, manifests.APPEdgeCorednsServiceGrid)
	edgeCorednsServiceGrid := ReadYaml(userCorednsServiceGrid, manifests.EdgeCorednsServiceGridYaml)
	data, err = kubeclient.ParseString(edgeCorednsServiceGrid, option)
	if err != nil {
		return err
	}

	serviceGrid := new(v1.ServiceGrid)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, serviceGrid); err != nil {
		return err
	}
	err = wait.PollImmediate(time.Second, 5*time.Minute, func() (bool, error) {
		err := serviceGrpupClinet.SuperedgeV1().ServiceGrids(constant.NamespaceEdgeSystem).Delete(context.TODO(), "", metav1.DeleteOptions{})
		if err != nil {
			klog.Warningf("Waiting deploy edge-coredns ServiceGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Delete %s success!", manifests.APPEdgeCorednsServiceGrid)

	return nil
}

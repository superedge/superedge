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
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests/edgex"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func DeployEdgex(client *kubernetes.Clientset, manifestsDir string, modules []bool) error {
	if err := EnsureEdgexNamespace(client); err != nil {
		return err
	}

	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgex,
	}

	configmap_name := "common-variables"
	if _, err := client.CoreV1().ConfigMaps(constant.NamespaceEdgex).Get(context.TODO(), configmap_name, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			userManifests := filepath.Join(manifestsDir, edgex.EDGEX_CONFIGMAP)
			edgexConfig := ReadYaml(userManifests, edgex.EDGEX_CONFIGMAP_YAML)
			err = kubeclient.CreateResourceWithFile(client, edgexConfig, option)
			if err != nil {
				klog.Errorf("Deploy %s fail, error: %v", edgex.EDGEX_CONFIGMAP, err)
				return err
			}
		}
	}

	var sername string
	var seryaml string
	for edgexModule, isTrue := range modules {
		if !isTrue {
			continue
		}
		switch edgexModule {
		case constant.App:
			sername = edgex.EDGEX_APP
			seryaml = edgex.EDGEX_APP_YAML
		case constant.Core:
			sername = edgex.EDGEX_CORE
			seryaml = edgex.EDGEX_CORE_YAML
		case constant.Support:
			sername = edgex.EDGEX_SUPPORT
			seryaml = edgex.EDGEX_SUPPORT_YAML
		case constant.Device:
			sername = edgex.EDGEX_DEVICE
			seryaml = edgex.EDGEX_DEVICE_YAML
		case constant.Sysmgmt:
			sername = edgex.EDGEX_SYS_MGMT
			seryaml = edgex.EDGEX_SYS_MGMT_YAML
		case constant.Ui:
			sername = edgex.EDGEX_UI
			seryaml = edgex.EDGEX_UI_YAML
		}
		klog.V(1).Infof("Start install %s to your cluster", sername)
		userManifests := filepath.Join(manifestsDir, sername)
		edgexYaml := ReadYaml(userManifests, seryaml)
		err := kubeclient.CreateResourceWithFile(client, edgexYaml, option)
		if err != nil {
			klog.Errorf("Deploy %s fail, error: %v", sername, err)
			return err
		}
		klog.V(1).Infof("Deploy %s success!", sername)
	}
	return nil
}

func DeployEdgeAPPS(client *kubernetes.Clientset, manifestsDir, caCertFile, caKeyFile, masterPublicAddr string, certSANs []string, configPath string) error {
	if err := EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}
	if err := DeployEdgePreflight(client, manifestsDir, masterPublicAddr, configPath); err != nil {
		return err
	}
	// Deploy tunnel
	if err := DeployTunnelAddon(client, manifestsDir, caCertFile, caKeyFile, masterPublicAddr, certSANs); err != nil {
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_EDGE)

	// Deploy edge-health
	if err := DeployEdgeHealth(client, manifestsDir); err != nil {
		klog.Errorf("Deploy edge health, error: %s", err)
		return err
	}
	klog.Infof("Deploy edge-health success!")

	// Deploy service-group
	if err := DeployServiceGroup(client, manifestsDir); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}
	klog.Infof("Deploy service-group success!")

	// Deploy edge-coredns
	if err := DeployEdgeCorednsAddon(configPath, manifestsDir); err != nil {
		klog.Errorf("Deploy edge-coredns error: %v", err)
		return err
	}

	// Update Kube-* Config
	if err := UpdateKubeConfig(client); err != nil {
		klog.Errorf("Deploy serivce group, error: %s", err)
		return err
	}
	klog.Infof("Update Kubernetes cluster config support marginal autonomy success")

	//Prepare config join Node
	if err := JoinNodePrepare(client, manifestsDir, caCertFile, caKeyFile); err != nil {
		klog.Errorf("Prepare config join Node error: %s", err)
		return err
	}
	klog.Infof("Prepare join Node configMap success")

	return nil
}

func DeleteEdgex(client *kubernetes.Clientset, manifestsDir string, modules []bool) error {
	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgex,
	}

	configmap_name := "common-variables"
	if _, err := client.CoreV1().ConfigMaps(constant.NamespaceEdgex).Get(context.TODO(), configmap_name, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			userManifests := filepath.Join(manifestsDir, edgex.EDGEX_CONFIGMAP)
			edgexConfig := ReadYaml(userManifests, edgex.EDGEX_CONFIGMAP_YAML)
			err = kubeclient.CreateResourceWithFile(client, edgexConfig, option)
			if err != nil {
				klog.Errorf("Deploy %s fail, error: %v", edgex.EDGEX_CONFIGMAP, err)
				return err
			}
		}
	}

	var sername string
	var seryaml string
	for edgexModule, isTrue := range modules {
		if !isTrue {
			continue
		}
		switch edgexModule {
		case constant.App:
			sername = edgex.EDGEX_APP
			seryaml = edgex.EDGEX_APP_YAML
		case constant.Core:
			sername = edgex.EDGEX_CORE
			seryaml = edgex.EDGEX_CORE_YAML
		case constant.Support:
			sername = edgex.EDGEX_SUPPORT
			seryaml = edgex.EDGEX_SUPPORT_YAML
		case constant.Device:
			sername = edgex.EDGEX_DEVICE
			seryaml = edgex.EDGEX_DEVICE_YAML
		case constant.Sysmgmt:
			sername = edgex.EDGEX_SYS_MGMT
			seryaml = edgex.EDGEX_SYS_MGMT_YAML
		case constant.Ui:
			sername = edgex.EDGEX_UI
			seryaml = edgex.EDGEX_UI_YAML
		case constant.Completely:
			sername = edgex.EDGEX_CONFIGMAP
			seryaml = edgex.EDGEX_CONFIGMAP_YAML
		}
		klog.V(1).Infof("Start uninstall %s from your cluster", sername)
		userManifests := filepath.Join(manifestsDir, sername)
		edgexYaml := ReadYaml(userManifests, seryaml)
		err := kubeclient.DeleteResourceWithFile(client, edgexYaml, option)
		if err != nil {
			klog.Errorf("Detach %s fail, error: %v", sername, err)
			return err
		}
		klog.V(1).Infof("Detach %s success!", sername)
	}

	if modules[constant.Completely] {
		klog.V(1).Infof("Start uninstall edgex completely.")
		err := os.RemoveAll("/consul")
		if err != nil {
			klog.Errorf("Delete /consul fail, err: %v.\nPlease 'rm /consul' by yourself.", err)
			return err
		}
		err = os.RemoveAll("/data")
		if err != nil {
			klog.Errorf("Delete /data fail, err: %v\nPlease 'rm /data' by yourself.", err)
			return err
		}
		klog.V(1).Infof("Uninstall edgex completely success!")
	}
	return nil
}

func DeleteEdgeAPPS(client *kubernetes.Clientset, manifestsDir, caCertFile, caKeyFile string, masterPublicAddr string, certSANs []string, configPath string) error {
	if ok := CheckIfEdgeAppDeletable(client); !ok {
		klog.Info("Can not Delete Edge Apps, cluster has remaining edge nodes!")
		return nil
	}

	// Deploy tunnel
	if err := DeleteTunnelAddon(client, manifestsDir, caCertFile, caKeyFile, masterPublicAddr, certSANs); err != nil {
		return err
	}
	klog.Infof("Delete %s success!", manifests.APP_TUNNEL_EDGE)

	// Delete edge-health
	if err := DeleteEdgeHealth(client, manifestsDir); err != nil {
		klog.Errorf("Delete edge health, error: %s", err)
		return err
	}
	klog.Infof("Delete edge-health success!")

	// Delete service-group
	if err := DeleteServiceGroup(client, manifestsDir); err != nil {
		klog.Errorf("Delete serivce group, error: %s", err)
		return err
	}
	klog.Infof("Delete service-group success!")

	// Delete edge-Coredns
	if err := DeleteEdgeCoredns(configPath, manifestsDir); err != nil {
		klog.Errorf("Delete edge-coredns, error: %s", err)
		return err
	}
	klog.Infof("Delete edge-Coredns success!")

	// Recover Kube-* Config
	if err := RecoverKubeConfig(client); err != nil {
		klog.Errorf("Recover Kubernetes cluster config support marginal autonomy, error: %s", err)
		return err
	}
	klog.Infof("Recover Kubernetes cluster config support marginal autonomy success")

	// Delete lite-api-server Cert
	if err := DeleteLiteApiServerCert(client); err != nil {
		klog.Errorf("Recover lite-apiserver, error: %s", err)
		return err
	}
	klog.Infof("Recover lite-apiserver configMap success")

	return nil
}

func ReadYaml(inputPath, defaults string) string {
	var yaml string = defaults
	if util.IsFileExist(inputPath) {
		yamlData, err := util.ReadFile(inputPath)
		if err != nil {
			klog.Errorf("Read yaml file: %s, error: %v", inputPath, err)
		}
		yaml = string(yamlData)
	}
	return yaml
}

func CreateByYamlFile(clientSet kubernetes.Interface, yamlFile string) error {
	err := kubeclient.CreateResourceWithFile(clientSet, yamlFile, nil)
	if err != nil {
		klog.Errorf("Apply yaml: %s, error: %v", yamlFile, err)
		return err
	}
	return nil
}

func DeleteByYamlFile(clientSet kubernetes.Interface, yamlFile string) error {
	err := kubeclient.DeleteResourceWithFile(clientSet, yamlFile, nil)
	if err != nil {
		klog.Errorf("Delete yaml: %s, error: %v", yamlFile, err)
		return err
	}
	return nil
}

func DeployHelperJob(clientSet *kubernetes.Clientset, helperYaml, action, role string) error {
	var err error
	var nodes *v1.NodeList
	var labelsNode = labels.NewSelector()

	if role == constant.NodeRoleNode {
		label := labels.SelectorFromSet(labels.Set(map[string]string{"app": "helper"}))
		if err := ClearJob(clientSet, label.String()); err != nil {
			return err
		}

		masterLabel, _ := labels.NewRequirement(constant.KubernetesDefaultRoleLabel, selection.NotIn, []string{""})
		changeLabel, _ := labels.NewRequirement(constant.EdgeChangeLabelKey, selection.Equals, []string{constant.EdgeChangeLabelValueEnable})
		nodeLabel, _ := labels.NewRequirement(constant.EdgeNodeLabelKey, selection.Equals, []string{constant.EdgeNodeLabelValueEnable})
		if action == constant.ActionChange {
			nodeLabel, _ = labels.NewRequirement(constant.EdgeNodeLabelKey, selection.NotIn, []string{constant.EdgeNodeLabelValueEnable})
		}

		labelsNode = labelsNode.Add(*masterLabel, *changeLabel, *nodeLabel)
		labelSelector := metav1.ListOptions{LabelSelector: labelsNode.String()}
		nodes, err = clientSet.CoreV1().Nodes().List(context.TODO(), labelSelector)
		if err != nil {
			return err
		}
	}

	if role == constant.NodeRoleMaster {
		masterLabel, _ := labels.NewRequirement(constant.KubernetesDefaultRoleLabel, selection.Equals, []string{""})
		masterNodeLabel, _ := labels.NewRequirement(constant.EdgeMasterLabelKey, selection.Equals, []string{constant.EdgeMasterLabelValueEnable})
		if action == constant.ActionChange {
			masterNodeLabel, _ = labels.NewRequirement(constant.EdgeMasterLabelKey, selection.NotIn, []string{constant.EdgeMasterLabelValueEnable})
		}

		labelsNode = labelsNode.Add(*masterLabel, *masterNodeLabel)
		labelSelector := metav1.ListOptions{LabelSelector: labelsNode.String()}
		nodes, err = clientSet.CoreV1().Nodes().List(context.TODO(), labelSelector)
		if err != nil {
			return err
		}
	}

	if action == constant.ActionChange {
		kubeclient.DeleteResourceWithFile(clientSet, manifests.HelperJobRbacYaml, "")
		time.Sleep(time.Second)

		option := map[string]interface{}{
			"Namespace": constant.NamespaceEdgeSystem,
		}
		if err := kubeclient.CreateResourceWithFile(clientSet, manifests.HelperJobRbacYaml, option); err != nil {
			return err
		}
	}

	kubeConf, err := util.ReadFile(os.Getenv("KUBECONFIG"))
	if err != nil {
		return err
	}

	masterIps, err := GetMasterIps(clientSet)
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		option := map[string]interface{}{
			"Namespace":  constant.NamespaceEdgeSystem,
			"NodeRole":   role,
			"Action":     action,
			"NodeName":   node.Name,
			"MasterIP":   masterIps[0],
			"KubeConfig": base64.StdEncoding.EncodeToString(kubeConf),
		}

		klog.V(4).Infof("Ready change node: %s", node.Name)
		if role == constant.NodeRoleNode {
			kubeclient.DeleteResourceWithFile(clientSet, helperYaml, option)

			time.Sleep(time.Duration(3) * time.Second)
			if err := kubeclient.CreateResourceWithFile(clientSet, helperYaml, option); err != nil {
				return err
			}
			continue
		}

		if role == constant.NodeRoleMaster {
			kubeclient.DeleteResourceWithFile(clientSet, helperYaml, option)

			time.Sleep(time.Duration(3) * time.Second)
			if err := kubeclient.CreateResourceWithFile(clientSet, helperYaml, option); err != nil {
				return err
			}
		}
	}

	fmt.Printf("Deploy helper-job-%s* success!\n", role)

	return nil
}

func GetMasterIps(clientSet kubernetes.Interface) ([]string, error) {
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var masterIPs []string
	for _, node := range nodes.Items {
		if _, ok := node.Labels[constant.KubernetesDefaultRoleLabel]; ok {
			for _, address := range node.Status.Addresses {
				if address.Type == v1.NodeInternalIP {
					masterIPs = append(masterIPs, address.Address)
				}
			}
		}
	}

	return masterIPs, nil
}

func ClearJob(clientSet *kubernetes.Clientset, label string) error {
	var gracePeriodSeconds int64 = 0
	jobOpts := metav1.ListOptions{
		LabelSelector: label,
	}
	jods, err := clientSet.BatchV1().Jobs(constant.NamespaceEdgeSystem).List(context.TODO(), jobOpts)
	if err != nil {
		return err
	}
	for _, job := range jods.Items {
		clientSet.BatchV1().Jobs(constant.NamespaceEdgeSystem).Delete(context.TODO(), job.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		})
	}

	podOpts := metav1.ListOptions{
		LabelSelector: label,
	}
	pods, err := clientSet.CoreV1().Pods(constant.NamespaceEdgeSystem).List(context.TODO(), podOpts)
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		clientSet.CoreV1().Pods(constant.NamespaceEdgeSystem).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		})
	}

	time.Sleep(time.Duration(3) * time.Second)

	return nil
}

func CheckIfEdgeAppDeletable(clientSet kubernetes.Interface) bool {
	nodeLabel, _ := labels.NewRequirement(constant.EdgeNodeLabelKey, selection.Equals, []string{constant.EdgeNodeLabelValueEnable})
	var labelsNode = labels.NewSelector()
	labelsNode = labelsNode.Add(*nodeLabel)
	labelSelector := metav1.ListOptions{LabelSelector: labelsNode.String()}
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), labelSelector)
	if err != nil {
		klog.Error(err)
		return false
	}

	if 0 == len(nodes.Items) {
		return true
	}
	return false
}

func EnsureEdgeSystemNamespace(client kubernetes.Interface) error {
	if err := kubeclient.CreateOrUpdateNamespace(client, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.NamespaceEdgeSystem,
		},
	}); err != nil {
		return err
	}
	return nil
}

func EnsureEdgexNamespace(client kubernetes.Interface) error {
	if err := kubeclient.CreateOrUpdateNamespace(client, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.NamespaceEdgex,
		},
	}); err != nil {
		return err
	}
	return nil
}

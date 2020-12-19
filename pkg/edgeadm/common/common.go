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
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"superedge/pkg/edgeadm/constant"
	"superedge/pkg/edgeadm/constant/manifests"
	"superedge/pkg/util"
	"superedge/pkg/util/kubeclient"
)

func ReadYaml(intputPath, defaults string) string {
	var yaml string = defaults
	if util.IsFileExist(intputPath) {
		yamlData, err := util.ReadFile(intputPath)
		if err != nil {
			klog.Errorf("Read yaml file: %s, error: %v", intputPath, err)
		}
		yaml = string(yamlData)
	}
	return yaml
}

func CreateByYamlFile(clientSet *kubernetes.Clientset, yamlFile string) error {
	err := kubeclient.CreateResourceWithFile(clientSet, yamlFile, nil)
	if err != nil {
		klog.Errorf("Apply yaml: %s, error: %v", yamlFile, err)
		return err
	}
	return nil
}

func DeleteByYamlFile(clientSet *kubernetes.Clientset, yamlFile string) error {
	err := kubeclient.DeleteResourceWithFile(clientSet, yamlFile, nil)
	if err != nil {
		klog.Errorf("Delete yaml: %s, error: %v", yamlFile, err)
		return err
	}
	return nil
}

func DeployHelperJob(clientSet *kubernetes.Clientset, helperYaml, action, role string) error {
	if role == constant.NODE_ROLE_NODE {
		label := labels.SelectorFromSet(labels.Set(map[string]string{"app": "helper"}))
		if err := ClearJob(clientSet, label.String()); err != nil {
			return err
		}
	}

	if action == constant.ACTION_CHANGE {
		kubeclient.DeleteResourceWithFile(clientSet, manifests.HelperJobRbacYaml, "")
		time.Sleep(time.Second)

		if err := kubeclient.CreateResourceWithFile(clientSet, manifests.HelperJobRbacYaml, ""); err != nil {
			return err
		}
	}

	kubeConf, err := util.ReadFile(os.Getenv("KUBECONF"))
	if err != nil {
		return err
	}

	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	masterIps, err := GetMasterIps(clientSet)
	if err != nil {
		return err
	}

	for _, node := range nodes.Items {
		option := map[string]interface{}{
			"NodeRole": role,
			"Action":   action,
			"NodeName": node.Name,
			"MasterIP": masterIps[0],
			"KubeConf": base64.StdEncoding.EncodeToString(kubeConf),
		}

		if _, ok := node.Labels[constant.KubernetesDefaultRoleLabel]; !ok && role == constant.NODE_ROLE_NODE {
			kubeclient.DeleteResourceWithFile(clientSet, helperYaml, option)

			time.Sleep(time.Duration(3) * time.Second)
			if err := kubeclient.CreateResourceWithFile(clientSet, helperYaml, option); err != nil {
				return err
			}
			continue
		}

		if _, ok := node.Labels[constant.KubernetesDefaultRoleLabel]; ok && role == constant.NODE_ROLE_MASTER {
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

func GetMasterIps(clientSet *kubernetes.Clientset) ([]string, error) {
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
	jods, err := clientSet.BatchV1().Jobs("kube-system").List(context.TODO(), jobOpts)
	if err != nil {
		return err
	}
	for _, job := range jods.Items {
		clientSet.BatchV1().Jobs("kube-system").Delete(context.TODO(), job.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		})
	}

	podOpts := metav1.ListOptions{
		LabelSelector: label,
	}
	pods, err := clientSet.CoreV1().Pods("kube-system").List(context.TODO(), podOpts)
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		clientSet.CoreV1().Pods("kube-system").Delete(context.TODO(), pod.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		})
	}

	time.Sleep(time.Duration(3) * time.Second)

	return nil
}

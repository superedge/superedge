/*
Copyright 2022 The SuperEdge Authors.

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

package e2e

import (
	"context"
	"fmt"
	crdClientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	siteClientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"testing"
	"time"
)

func initClient(t *testing.T) (*kubernetes.Clientset, *crdClientset.Clientset, *siteClientset.Clientset) {
	config, err := clientcmd.BuildConfigFromFlags("", "/tmp/config")
	if err != nil {
		t.Fatal("can not init build k8s config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatal("can not init clientset")
	}
	crdClient := crdClientset.NewForConfigOrDie(config)
	siteClient := siteClientset.NewForConfigOrDie(config)
	return clientset, crdClient, siteClient
}

func waitForDeployment(t *testing.T, ctx context.Context, namespace string, deploymentname string, clientSet *kubernetes.Clientset) (bool, error) {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("Timeout, still restart count not as expected")
			return false, fmt.Errorf("timeout Error")
		case <-tick:
			_, err := clientSet.AppsV1().Deployments(namespace).Get(ctx, deploymentname, metav1.GetOptions{})
			if err == nil {
				t.Log("Deployment get as expected")
				return true, nil
			}
		}
	}
}

func waitForSVC(t *testing.T, ctx context.Context, namespace string, svcname string, clientSet *kubernetes.Clientset) (bool, error) {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("Timeout, still restart count not as expected")
			return false, fmt.Errorf("timeout Error")
		case <-tick:
			_, err := clientSet.CoreV1().Services(namespace).Get(ctx, svcname, metav1.GetOptions{})
			if err == nil {
				t.Log("Service get as expected")
				return true, nil
			}
		}
	}
}
func waitForNodeUnit(t *testing.T, ctx context.Context, name string, crdClient *siteClientset.Clientset) (bool, error) {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			t.Log("Timeout, still restart count not as expected")
			return false, fmt.Errorf("timeout Error")
		case <-tick:
			_, err := crdClient.SiteV1alpha1().NodeGroups().Get(ctx, name, metav1.GetOptions{})
			if err == nil {
				t.Log("nodeunit get as expected")
				return true, nil
			}
		}
	}
}

func ReadyNodes(ctx context.Context, client kubernetes.Interface, nodeInformer coreinformers.NodeInformer, nodeSelector string) ([]*v1.Node, error) {
	ns, err := labels.Parse(nodeSelector)
	if err != nil {
		return []*v1.Node{}, err
	}

	var nodes []*v1.Node
	// err is defined above
	if nodes, err = nodeInformer.Lister().List(ns); err != nil {
		return []*v1.Node{}, err
	}

	if len(nodes) == 0 {

		nItems, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: nodeSelector})
		if err != nil {
			return []*v1.Node{}, err
		}

		if nItems == nil || len(nItems.Items) == 0 {
			return []*v1.Node{}, nil
		}

		for i := range nItems.Items {
			node := nItems.Items[i]
			nodes = append(nodes, &node)
		}
	}

	readyNodes := make([]*v1.Node, 0, len(nodes))
	for _, node := range nodes {
		if IsReady(node) {
			readyNodes = append(readyNodes, node)
		}
	}
	return readyNodes, nil
}

// IsReady checks if the superedge could run against given node.
func IsReady(node *v1.Node) bool {
	for i := range node.Status.Conditions {
		cond := &node.Status.Conditions[i]
		if cond.Type == v1.NodeReady && cond.Status != v1.ConditionTrue {

			return false
		}
	}

	return true
}

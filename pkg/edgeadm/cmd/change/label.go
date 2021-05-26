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

package change

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

type labelAction struct {
	all       bool
	clientSet *kubernetes.Clientset
	flags     *pflag.FlagSet
	selector  string
	nodeInfos *v1.NodeList
}

func newLabel() labelAction {
	return labelAction{}
}

func newLabelCMD() *cobra.Command {
	action := newLabel()
	cmd := &cobra.Command{
		Use:   "label [NODENAME]",
		Short: "Label the node to be change.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := action.complete(cmd, args); err != nil {
				util.OutPutMessage(err.Error())
				return
			}

			if err := action.runLabel(); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}

	action.flags = cmd.Flags()
	cmd.Flags().StringVarP(&action.selector, "selector", "l", action.selector, "Use label to select the node that needs to be changed")

	return cmd
}

func (l *labelAction) complete(cmd *cobra.Command, args []string) error {
	if err := l.checkKubeConfig(); err != nil {
		return err
	}
	if len(args) > 0 && (len(l.selector) > 0) {
		return fmt.Errorf("Cannot specify both a nodename or --selector option")
	}

	l.nodeInfos = &v1.NodeList{}
	if len(args) > 0 {
		for _, name := range args {
			node, err := l.clientSet.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			l.nodeInfos.Items = append(l.nodeInfos.Items, *node)
		}
	}

	if len(l.selector) > 0 || len(args) == 0 {
		opts := metav1.ListOptions{
			LabelSelector: l.selector,
		}
		nodeList, err := l.clientSet.CoreV1().Nodes().List(context.TODO(), opts)
		if err != nil {
			return err
		}
		l.nodeInfos = nodeList
	}

	return nil
}

func (l *labelAction) checkKubeConfig() error {
	configPath, err := l.flags.GetString("kubeconfig")
	if err != nil {
		klog.Errorf("Get kubeconfig flags error: %v", err)
	}

	l.clientSet, err = kubeclient.GetClientSet(configPath)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}
	if l.clientSet == nil {
		return fmt.Errorf("Please set kubeconfig value!\n")
	}

	return nil
}

func (l *labelAction) runLabel() error {
	for _, node := range l.nodeInfos.Items {
		if err := l.labelToEdgeNode(&node); err != nil {
			return fmt.Errorf("error: unable to add edge label to the node %q: %v\n", node.Name, err)
		}
	}
	return nil
}

func (l *labelAction) labelToEdgeNode(node *v1.Node) error {
	patchData := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"labels": {
				constant.EdgeChangeLabelKey: constant.EdgeChangeLabelValueEnable,
			},
		},
	}
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return err
	}

	if _, err := l.clientSet.CoreV1().Nodes().Patch(context.TODO(), node.Name, types.StrategicMergePatchType,
		patchBytes, metav1.PatchOptions{}); err != nil {
		return err
	}
	return nil
}

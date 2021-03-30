package steps

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"net/http"
)

func NewEdgeKubeletPhase() workflow.Phase {
	return workflow.Phase{
		Name:    "edge kubelet",
		Short:   "Install edge kubelet",
		Long:    "Install edge kubelet",
		Example: dockerExample,
		Run:     installEdgeKubelet,
		InheritFlags: []string{
			options.CfgPath,               //todo
			options.IgnorePreflightErrors, //todo
		},
	}
}

// runPreflight executes preflight checks logic.
func installEdgeKubelet(c workflow.RunData) error {
	// Deploy edge kubelet
	klog.Infof("Node: %s Start deploy edge kubelet", options.NodeName)
	if err := deployKubelet(options.NodeName); err != nil {
		klog.Errorf("Restart kubelet error: %v", err)
		return err
	}
	return nil
}

func deployKubelet(nodeName string) error {
	if err := updateKubeletConfig(constant.KubeadmKubeletConfig, constant.EdgeadmKubeletConfig); err != nil {
		klog.Errorf("Update kubelet config error: %v", err)
		return err
	}

	if err := util.WriteWithBufio(constant.KubeletStartEnvFile, constant.CHANGE_KUBELET_KUBECONFIG_ARGS); err != nil {
		klog.Errorf("Write kubelet start env file: %s error: %v", constant.KubeletStartEnvFile, err)
		return err
	}
	klog.Infof("Node: %s update kubelet config success.", nodeName)

	if _, _, err := util.RunLinuxCommand(constant.KUBELET_RESTART_CMD); err != nil {
		klog.Errorf("Running linux command: %s error: %v", constant.KUBELET_RESTART_CMD, err)
		return err
	}
	klog.Infof("Node: %s Restart kubelet config success.", nodeName)

	if _, _, err := util.RunLinuxCommand(constant.KUBELET_STATUS_CMD); err != nil {
		klog.Errorf("Running linux command: %s error: %v", constant.KUBELET_RESTART_CMD, err)
		return err
	}
	klog.Infof("Node: %s Status kubelet config success.", nodeName)

	if err := checkKubletHealthz(); err != nil {
		return fmt.Errorf("Node: %s is NotReady, error: %v\n", nodeName, err)
	}

	//Check link health using kubelet client request kube-api-server by lite-apiserver
	kubeletCleint, err := kubeclient.GetClientSet(constant.EdgeadmKubeletConfig)
	if err != nil {
		return err
	}
	isReady, err := isRunningKubelet(kubeletCleint, nodeName, 60)
	if err != nil {
		klog.Errorf("Check kubelet status error: %v", err)
		return err
	}

	if !isReady {
		return fmt.Errorf("Node: %s is NotReady\n", nodeName)
	}

	if err := addLiteFinishLabel(kubeletCleint, nodeName); err != nil {
		klog.Errorf("Add LiteApiServer Running label error: %v", err)
		return err
	}
	klog.Infof("Node: %s success deploy lite-apiserver.", nodeName)

	return nil
}

func isRunningKubelet(kubeClient *kubernetes.Clientset, nodeName string, retry int) (bool, error) {
	nodeStatueFunc := func(nodeName string) (bool, error) {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get node: %s infos error: %v", nodeName, err)
			return false, err
		}

		for i := range node.Status.Conditions {
			if node.Status.Conditions[i].Type == v1.NodeReady &&
				node.Status.Conditions[i].Status == v1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	}

	for i := 0; i < retry; i++ {
		time.Sleep(time.Second)
		isRunning, err := nodeStatueFunc(nodeName)
		if err != nil {
			klog.Errorf("Get node: %s status infos error: %v", nodeName, err)
			continue
		}
		if isRunning {
			klog.Infof("Check kubelet already running after restart kubelet")
			return true, nil
		}
	}

	return false, nil
}

func updateKubeletConfig(srcFile, dstFile string) error {
	config, err := clientcmd.LoadFromFile(srcFile)
	if err != nil {
		return err
	}

	for key := range config.Clusters {
		config.Clusters[key].Server = constant.LiteAPIServerAddr
	}

	if err = clientcmd.WriteToFile(*config, dstFile); err != nil {
		return err
	}

	return nil
}

func checkKubletHealthz() error {
	return wait.PollImmediate(time.Second, 3*time.Minute, func() (bool, error) {
		resp, err := http.Get(constant.KubeletHealthzURl)
		if err != nil {
			return false, nil
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		klog.Infof("Check kubelet healthz get resp: %s body: %s", util.ToJson(resp), util.ToJson(body))

		return resp.StatusCode == http.StatusOK, nil
	})
}

func addLiteFinishLabel(kubeClient *kubernetes.Clientset, nodeName string) error {
	return wait.PollImmediate(time.Second, 3*time.Minute, func() (bool, error) {
		node, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get node: %s error: %v", nodeName, err)
			return false, nil
		}

		node.ObjectMeta.Labels[constant.EDGE_NODE_KEY] = nodeName
		if _, err := kubeClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("Update node: %s labels error: %v", nodeName, err)
			return false, nil
		}
		return true, nil
	})
}

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

package deploy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/superedge/superedge/pkg/helper-job/common"
	"github.com/superedge/superedge/pkg/helper-job/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func changeMasterJob(kubeClient *kubernetes.Clientset, nodeName string) error {
	klog.Infof("Init Master %s Start.", nodeName)
	if err := addMasterHosts(kubeClient); err != nil {
		klog.Errorf("Add master hosts error: %v", err)
		return err
	}

	masterLabel := map[string]string{
		util.EdgeMasterLabelKey: util.EdgeMasterLabelValueEnable,
	}
	isLabel, err := kubeclient.CheckNodeLabel(kubeClient, nodeName, masterLabel)
	if err != nil {
		klog.Errorf("Check is deploy LiteAPiServer error: %v", err)
		return err
	}
	if !isLabel {
		if err := writeKubeAPIYaml(kubeClient,
			constant.KubeAPIServerSrcYamlPath, constant.KubeAPIServerBackUPYamlPath); err != nil {
			return err
		}

		if err := kubeclient.AddNodeLabel(kubeClient, nodeName, masterLabel); err != nil {
			klog.Errorf("Add edged master node label error: %v", err)
			return err
		}
	}

	klog.Infof("Init Master %s Success.", nodeName)
	return nil
}

func changeNodeJob(kubeClient *kubernetes.Clientset, nodeName string) error {
	// Deploy LiteAPIServer
	klog.Infof("Node: %s Start deploy LiteAPIServer", nodeName)
	isDeploy, err := isDeployLiteAPIServer(kubeClient, nodeName)
	if isDeploy || err != nil {
		return err
	}

	if err := deployLiteAPIServer(kubeClient, nodeName); err != nil {
		klog.Errorf("Deploy LiteAPIServer error: %v", err)
		return err
	}

	if err := restartKubelet(kubeClient, nodeName); err != nil {
		klog.Errorf("Restart kubelet error: %v", err)
		return err
	}

	return nil
}

func isDeployLiteAPIServer(kubeClient *kubernetes.Clientset, nodeName string) (bool, error) {
	nodeLabel := map[string]string{
		util.EdgeNodeLabelKey: util.EdgeNodeLabelValueEnable,
	}
	isLabel, err := kubeclient.CheckNodeLabel(kubeClient, nodeName, nodeLabel)
	if err != nil {
		klog.Errorf("Check is deploy LiteAPiServer error: %v", err)
		return false, err
	}

	isRunning, err := isRunningLiteAPIServer(kubeClient, nodeName, 3)
	if err != nil {
		klog.Errorf("Check is Running of LiteAPiServer before deploy error: %v", err)
		return false, err
	}

	if isLabel && isRunning {
		klog.Infof("Node: %s LiteAPIServer is healthy, deploy finish!", nodeName)
		return true, nil
	}

	if !isRunning && isLabel {
		if err := kubeclient.DeleteNodeLabel(kubeClient, nodeName, nodeLabel); err != nil {
			return false, err
		}
	}

	return false, nil
}

func restartKubelet(kubeClient *kubernetes.Clientset, nodeName string) error {
	if err := updateKubeletConfig(constant.KubeadmKubeletConfig, constant.EdgeadmKubeletConfig); err != nil {
		klog.Errorf("Update kubelet config error: %v", err)
		return err
	}

	if err := util.WriteWithBufio(constant.KubeletStartEnvFile, constant.CHANGE_KUBELET_KUBECONFIG_ARGS); err != nil {
		klog.Errorf("Write kubelet start env file: %s error: %v", constant.KubeletStartEnvFile, err)
		return err
	}
	klog.Infof("Node: %s update kubelet config success.", nodeName)

	if err := runLinuxCommand(constant.KUBELET_RESTART_CMD); err != nil {
		klog.Errorf("Running linux command: %s error: %v", constant.KUBELET_RESTART_CMD, err)
		return err
	}
	klog.Infof("Node: %s Restart kubelet config success.", nodeName)

	if err := runLinuxCommand(constant.KUBELET_STATUS_CMD); err != nil {
		klog.Errorf("Running linux command: %s error: %v", constant.KUBELET_RESTART_CMD, err)
		return err
	}
	klog.Infof("Node: %s Status kubelet config success.", nodeName)

	if err := checkKubeletHealthz(); err != nil {
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

	masterLabel := map[string]string{
		util.EdgeNodeLabelKey: util.EdgeNodeLabelValueEnable,
	}
	if err := kubeclient.AddNodeLabel(kubeClient, nodeName, masterLabel); err != nil {
		klog.Errorf("Add edged Node node label error: %v", err)
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

func updateKubeAPIPod(kubeClient *kubernetes.Clientset, pod *v1.Pod) error {
	if len(pod.Spec.Containers) < 1 {
		return errors.New("Get kube-api-server container nil\n")
	}

	commands := pod.Spec.Containers[0].Command
	for index, cmd := range commands {
		if strings.Contains(cmd, "--kubelet-preferred-address-types=") {
			commands[index] = "--kubelet-preferred-address-types=Hostname"
			break
		}
	}
	pod.Spec.Containers[0].Command = commands

	var clusterIP string
	for { //make sure tunnel-coredns success created
		coredns, err := kubeClient.CoreV1().Services(util.NamespaceEdgeSystem).Get(context.TODO(), "tunnel-coredns", metav1.GetOptions{})
		if err == nil {
			clusterIP = coredns.Spec.ClusterIP
			break
		}
		time.Sleep(time.Second)
	}
	if clusterIP == "" {
		return errors.New("Get tunnel-coredns ClusterIP nil\n")
	}

	if pod.Spec.DNSConfig == nil {
		podDNS := &v1.PodDNSConfig{
			Nameservers: []string{clusterIP},
		}
		pod.Spec.DNSConfig = podDNS
	} else {
		pod.Spec.DNSConfig.Nameservers = []string{clusterIP}
	}
	pod.Spec.DNSPolicy = "None"

	fmt.Printf("%s", util.ToJsonForm(pod))

	return nil
}

func writeKubeAPIYaml(kubeClient *kubernetes.Clientset, srcFile, dstFile string) error {
	yamlByte, err := util.ReadFile(srcFile)
	if err != nil {
		return err
	}
	util.WriteWithBufio(dstFile, string(yamlByte))

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(yamlByte, nil, nil)
	if err != nil {
		return err
	}
	pod := obj.(*v1.Pod)
	klog.Infof("Load master kube-api-service static yaml success.")

	if err := updateKubeAPIPod(kubeClient, pod); err != nil {
		return err
	}
	klog.Infof("Update master kube-api-service static yaml success.")

	podJson, err := json.Marshal(pod)
	if err != nil {
		return nil
	}
	data, err := k8syaml.JSONToYAML(podJson)
	if err != nil {
		return err
	}
	util.WriteWithBufio(srcFile, string(data))

	return nil
}

func addMasterHosts(kubeClient *kubernetes.Clientset) error {
	hosts, err := util.ReadFile(constant.MasterHostsFilePath)
	if err != nil {
		return err
	}

	masterIps, err := common.GetMasterHosts(kubeClient)
	if err != nil {
		return err
	}

	content := ""
	hostSrt := string(hosts)
	for hostname := range masterIps {
		host := fmt.Sprintf("%s %s \n", masterIps[hostname], hostname)
		if strings.Contains(hostSrt, host) {
			continue
		}
		content += host
	}

	return util.WriteWithAppend(constant.MasterHostsFilePath, content)
}

func runLinuxCommand(command string) error {
	var outBuff bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", command)

	cmd.Stdout = &outBuff
	cmd.Stderr = &outBuff
	defer func() {
		defer klog.Infof("Get command: '%s' output: \n %s", command, outBuff.String())
	}()

	//Run cmd
	if err := cmd.Start(); err != nil {
		klog.Errorf("Exec command: %s, error: %v", command, err)
		return err
	}

	//Wait cmd run finish
	if err := cmd.Wait(); err != nil {
		klog.Errorf("Wait command: %s exec finish error: %v", command, err)
		return err
	}

	return nil
}

func deployLiteAPIServer(kubeClient *kubernetes.Clientset, nodeName string) error {
	if err := completeLiteApiServer(kubeClient); err != nil {
		klog.Errorf("Complete lite api server error: %v", err)
		return err
	}

	liteAPIServerTemplate, err := util.ReadFile(constant.LiteAPIServerTemplatePath)
	if err != nil {
		klog.Errorf("Read yaml file: %s, error: %v", constant.LiteAPIServerTemplatePath, err)
		return err
	}
	liteAPIServerYaml := string(liteAPIServerTemplate)

	masterIP := os.Getenv("MASTER_IP")
	if nodeName == "" {
		return errors.New("Get ENV MASTER_IP nil\n")
	}

	option := map[string]interface{}{
		"MasterIP": masterIP,
	}
	liteApiServerYaml, err := kubeclient.CompleteTemplate(liteAPIServerYaml, option)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(constant.KubeadmKubeletManifests, os.ModePerm); err != nil {
		return err
	}
	if err := writeStaticYaml(constant.LiteAPIServerYamlPath, liteApiServerYaml); err != nil {
		klog.Errorf("Write static yaml file: %s error: %v", constant.LiteAPIServerYamlPath, err)
		return err
	}
	klog.Infof("Node: %s Write LiteAPIServer yaml file success.", nodeName)

	return nil
}

func completeLiteApiServer(kubeClient *kubernetes.Clientset) error {
	edgeCertCM, err := kubeClient.CoreV1().ConfigMaps(util.NamespaceEdgeSystem).Get(context.TODO(), constant.EDGE_CERT_CM, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := os.RemoveAll(constant.KubeadmKubeletEdgeCert); err != nil {
		return err
	}

	if err := os.MkdirAll(constant.KubeadmKubeletEdgeCert, os.ModePerm); err != nil {
		return err
	}

	for key := range edgeCertCM.Data {
		if err := util.WriteWithBufio(constant.KubeadmKubeletEdgeCert+key, string(edgeCertCM.Data[key])); err != nil {
			klog.Errorf("Write file: %s, error: %v", key, err)
			return err
		}
	}

	return nil
}

func checkKubeletHealthz() error {
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

func writeStaticYaml(filePath, content string) error {
	var err error
	for i := 0; i < 3; i++ {
		if util.IsFileExist(filePath) {
			if err = os.Remove(filePath); err != nil {
				klog.Errorf("Delete file: %s error: %v", filePath, err)
				i--
				time.Sleep(time.Second)
				continue
			}
		}

		if err = util.WriteWithBufio(filePath, content); err != nil {
			klog.Errorf("Write file: %s error: %v", filePath, err)
			continue
		}
		return nil
	}

	return err
}

func isRunningLiteAPIServer(kubeClient *kubernetes.Clientset, nodeName string, retry int) (bool, error) {
	podName := constant.LiteAPIServerPodName + "-" + nodeName
	liteAPIServerStatueFunc := func(podName string) (bool, error) {
		pod, err := kubeClient.CoreV1().Pods(util.NamespaceEdgeSystem).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get pod: %s infos error: %v", podName, err)
			return false, err
		}

		if pod.Status.Phase == v1.PodRunning {
			return true, nil
		}

		return false, nil
	}

	for i := 0; i < retry; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		isRunning, err := liteAPIServerStatueFunc(podName)
		if err != nil {
			klog.Errorf("Get pod: %s status infos error: %v", podName, err)
			continue
		}
		if isRunning {
			return true, nil
		}
	}

	return false, nil
}

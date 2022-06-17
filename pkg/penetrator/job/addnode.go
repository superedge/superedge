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

package job

import (
	"context"
	"fmt"
	"github.com/superedge/superedge/pkg/penetrator/constants"
	"github.com/superedge/superedge/pkg/penetrator/job/conf"
	penetratorutil "github.com/superedge/superedge/pkg/penetrator/util"
	"github.com/superedge/superedge/pkg/util"
	kubeutil "github.com/superedge/superedge/pkg/util/kubeclient"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	//"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

func AddNodes(nodes int) {

	// Control the number of concurrently added nodes
	addch := make(chan interface{}, nodes)

	//Cache IP about failed addition of nodes
	errch := make(chan string, len(conf.JobConf.NodesIps))

	// Used to report event events
	nodejob := &batchv1.Job{}
	nodejob.Namespace = os.Getenv(constants.JobNameSpace)
	nodejob.Name = os.Getenv(constants.JobName)

	//Get the kubeclient of the cluster
	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("failed to build in-cluster kubeconfig, error: %v", err)
		return
	}
	kubeclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to  build kubernetes clientset: %v", err)
		return
	}

	version, err := kubeclient.ServerVersion()
	if err != nil {
		klog.Errorf("fialed to get kubernetes version, error: %v", err)
		return
	}

	// Report job event
	userBoardcaster := record.NewBroadcaster()
	userBoardcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclient.CoreV1().Events(constants.NameSpaceEdge)})
	userBoardcaster.StartLogging(klog.Infof)
	userRecord := userBoardcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: os.Getenv(constants.JobName)})

	//apiserver advertiseAddress

	for node, ip := range conf.JobConf.NodesIps {
		go func() {
			err := addNode(node, ip, version.GitVersion, conf.JobConf.ApiserverAddr, addch, errch, kubeclient)
			if err != nil {
				klog.Error(err.Error())
				userRecord.Event(nodejob, v1.EventTypeWarning, fmt.Sprintf("Node:%s installation failed", node), err.Error())
			}
		}()
		addch <- struct{}{}
	}

	for {
		if len(addch) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if len(errch) != 0 {
		errNodes := make([]string, len(errch))
		for i := 0; i < len(errch); i++ {
			errNodes[i] = <-errch
		}
		userRecord.Event(nodejob, v1.EventTypeWarning, "Node installation failed", fmt.Sprintf("Node ips that was not installed successfully, ips: %v", errNodes))
		os.Exit(1)
	}

	klog.Infof("addnodes job complete !")
}

func addNode(nodeName, nodeIp, version, advertiseAddress string, nodesch chan interface{}, errNodech chan string, kubeclient kubernetes.Interface) error {
	defer func() {
		// Decrease the count of concurrently added nodes by one
		<-nodesch
	}()

	node, err := kubeclient.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get node:%s, error: %v", nodeName, err)
		}
	} else {
		if node.Labels[constants.NodeLabel] == conf.JobConf.NodeLabel {
			return nil
		}
		err = kubeclient.CoreV1().Nodes().Delete(context.Background(), node.Name, metav1.DeleteOptions{})
		if err != nil {
			errNodech <- nodeIp
			return fmt.Errorf("failed to delete node:%s, error: %v", nodeName, err)
		}

	}

	defer func() {
		errNodech <- nodeIp
	}()

	client, err := penetratorutil.SShConnectNode(nodeIp, conf.JobConf.SSHPort, conf.JobConf.Secret)
	if err != nil {
		return fmt.Errorf("failed to get ssh client, error: %v", err)
	}
	defer client.Close()

	archSession, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to get ssh client session, error: %v", err)
	}

	arch, err := archSession.CombinedOutput("uname -m")
	if err != nil {
		return fmt.Errorf("failed to get arch, error: %v", err)
	}

	simpleArch := getArch(string(arch))
	if simpleArch == "" {
		return fmt.Errorf("Unsupported arch %s", string(arch))
	}

	err = penetratorutil.ScpFile(nodeIp, fmt.Sprintf(constants.InstallPackage+"%s-%s.tar.gz", simpleArch, version), conf.JobConf.SSHPort, conf.JobConf.Secret)
	if err != nil {
		return fmt.Errorf("Failed to copy installation package, error: %v", err)
	}

	//Get the script for adding nodes
	option := map[string]interface{}{
		"NodeName":         nodeName,
		"CaHash":           conf.JobConf.CaHash,
		"AdvertiseAddress": advertiseAddress,
		"BindPort":         conf.JobConf.ApiserverPort,
		"AdmToken":         conf.JobConf.AdmToken,
		"Arch":             simpleArch,
		"K8sVersion":       version,
	}

	scriptTmep, err := util.ReadFile(constants.AddNodeScript)
	if err != nil {
		return fmt.Errorf("Failed to read file:%s, error: %v", constants.AddNodeScript, err)
	}
	script, err := kubeutil.CompleteTemplate(string(scriptTmep), option)
	if err != nil {
		return fmt.Errorf("Failed to get addnode.sh, error: %v", err)
	}

	scriptSession, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to get ssh client session, error: %v", err)
	}

	stdout, err := scriptSession.CombinedOutput(script)
	if err != nil {
		return fmt.Errorf("Failed to add node, error: %v, script: %s, stdout: %s", err, script, string(stdout))
	}

	err = kubeutil.AddNodeLabel(kubeclient, nodeName, map[string]string{constants.NodeLabel: conf.JobConf.NodeLabel})
	if err != nil {
		return fmt.Errorf("Failed to label node %s, error: %v", nodeName, err)
	}

	klog.Infof("Add node: %s successfully", nodeName)
	return nil
}

func getArch(arch string) string {
	arch = strings.Replace(arch, "\n", "", -1)
	switch arch {
	case constants.X86_64:
		return constants.Amd64
	case constants.Aarch64:
		return constants.Arm64
	default:
		return ""
	}

}

package steps

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"strings"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var (
	workPath = ""
	masterIP = ""
)

func NewLiteApiServerInitPhase(edgeMasterIP string, edgeWorkPath string) workflow.Phase {
	workPath = edgeWorkPath
	masterIP = edgeMasterIP
	return workflow.Phase{
		Name:  "lite-apiserver init",
		Short: "Install lite-apiserver on edge node",
		Long:  "Install lite-apiserver on edge node",
		Run:   installLiteAPIServer,
		InheritFlags: []string{
			options.CfgPath,               //todo
			options.IgnorePreflightErrors, //todo
		},
	}
}

// runPreflight executes preflight checks logic.
func installLiteAPIServer(c workflow.RunData) error {
	// Deploy LiteAPIServer
	klog.Infof("Node: %s Start deploy LiteAPIServer", options.NodeName)
	kubeClient, err := kubeclient.GetClientSet("")
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return err
	}
	isDeploy, err := isRunningLiteAPIServer()
	if isDeploy || err != nil {
		return err
	}
	if err := deployLiteAPIServer(kubeClient, options.NodeName); err != nil {
		klog.Errorf("Deploy LiteAPIServer error: %v", err)
		return err
	}

	return nil
}

func isRunningLiteAPIServer() (bool, error) {
	cmd := fmt.Sprintf(constant.LITE_APISERVER_STATUS_CMD)
	if _, _, err := util.RunLinuxCommand(cmd); err != nil {
		klog.Errorf("Running linux command: %s error: %v", cmd, err)
		return false, nil
	}
	return true, nil
}

func deployLiteAPIServer(kubeClient *kubernetes.Clientset, nodeName string) error {
	kubeService, err := kubeClient.CoreV1().Services(constant.NAMESPACE_DEFAULT).Get(context.TODO(), constant.SERVICE_KUBERNETES, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if kubeService.Spec.ClusterIP == "" {
		return errors.New("Get kubernetes service clusterIP nil\n")
	}

	generateLiteApiserverKey()
	generateLiteApiserverCsr(kubeService.Spec.ClusterIP)
	generateLiteApiserverTlsJson()
	createLiteApiserverConfig()
	startLiteApiserver()
	return nil
}

func generateLiteApiserverKey() error {
	cmds := []string{
		fmt.Sprintf("mkdir -p /etc/kubernetes/edge/ && openssl genrsa -out /etc/kubernetes/edge/lite-apiserver.key 2048"),
		fmt.Sprintf("cp /etc/kubernetes/edge/lite-apiserver.key /etc/kubernetes/pki/lite-apiserver.key"),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func generateLiteApiserverCsr(clusterIP string) error {
	cmds := []string{
		fmt.Sprintf("mkdir -p /etc/kubernetes/edge/ && cat << EOF >/etc/kubernetes/edge/lite-apiserver.conf\n[req]\ndistinguished_name = req_distinguished_name\nreq_extensions = v3_req\n[req_distinguished_name]\nCN = lite-apiserver\n[v3_req]\nbasicConstraints = CA:FALSE\nkeyUsage = nonRepudiation, digitalSignature, keyEncipherment\nsubjectAltName = @alt_names\n[alt_names]\nDNS.1 = localhost\nIP.1 = 127.0.0.1\nIP.2 = %s\nEOF", clusterIP),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl req -new -key lite-apiserver.key -subj \"/CN=lite-apiserver\" -config lite-apiserver.conf -out lite-apiserver.csr"),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl x509 -req -in lite-apiserver.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -days 5000 -extensions v3_req -extfile lite-apiserver.conf -out lite-apiserver.crt"),
		fmt.Sprintf("cp /etc/kubernetes/edge/lite-apiserver.crt /etc/kubernetes/pki/lite-apiserver.crt"),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func generateLiteApiserverTlsJson() error {
	cmds := []string{
		fmt.Sprintf("mkdir -p /etc/kubernetes/edge/ && cat << EOF >/etc/kubernetes/edge/tls.json\n[\n    {\n        \"key\":\"/var/lib/kubelet/pki/kubelet-client-current.pem\",\n        \"cert\":\"/var/lib/kubelet/pki/kubelet-client-current.pem\"\n    }\n]\nEOF"),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl req -new -key lite-apiserver.key -subj \"/CN=lite-apiserver\" -config lite-apiserver.conf -out lite-apiserver.csr"),
		fmt.Sprintf("cd /etc/kubernetes/edge/ && openssl x509 -req -in lite-apiserver.csr -CA /etc/kubernetes/pki/ca.crt -CAkey /etc/kubernetes/pki/ca.key -CAcreateserial -days 5000 -extensions v3_req -extfile lite-apiserver.conf -out lite-apiserver.crt"),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func createLiteApiserverConfig() error {
	liteApiserverConfigTemplate := constant.LiteApiserverTemplate
	strings.ReplaceAll(liteApiserverConfigTemplate, "${MASTER_IP}", masterIP)
	cmds := []string{
		fmt.Sprintf(`echo "%s" > %s`, liteApiserverConfigTemplate, constant.LiteApiserverConfPath),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func startLiteApiserver() error {
	cmds := []string{
		fmt.Sprintf(`cp %s %s`, constant.LiteApiserverBinPath, constant.UsrLocalBinDir),
		fmt.Sprintf(constant.LITE_APISERVER_RESTART_CMD),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

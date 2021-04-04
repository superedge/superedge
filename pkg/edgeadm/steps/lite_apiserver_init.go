package steps

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	phases "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/join"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	kubeconfigutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util/kubeconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

var (
	workPath = ""
	masterIP = ""
)

func NewLiteApiServerInitPhase(edgeWorkPath string) workflow.Phase {
	workPath = edgeWorkPath
	return workflow.Phase{
		Name:  "lite-apiserver init",
		Short: "Install lite-apiserver on edge node",
		Long:  "Install lite-apiserver on edge node",
		Run:   installLiteAPIServer,
		InheritFlags: []string{
			options.IgnorePreflightErrors, //todo
			options.CfgPath,
			options.NodeCRISocket,
			options.NodeName,
			options.FileDiscovery,
			options.TokenDiscovery,
			options.TokenDiscoveryCAHash,
			options.TokenDiscoverySkipCAHash,
			options.TLSBootstrapToken,
			options.TokenStr,
		},
	}
}

// runPreflight executes preflight checks logic.
func installLiteAPIServer(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("installLiteAPIServer phase invoked with an invalid data struct")
	}

	if data.Cfg().ControlPlane != nil {
		return nil
	}

	// Deploy LiteAPIServer
	klog.Infof("Node: %s Start deploy LiteAPIServer", options.NodeName)
	isDeploy, err := isRunningLiteAPIServer()
	if isDeploy || err != nil {
		return err
	}

	kubeClient, err := initKubeClient(data)
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return err
	}
	// Deletes the bootstrapKubeConfigFile, so the credential used for TLS bootstrap is removed from disk
	defer func() {
		os.Remove(kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
		os.Remove(constant.KubeadmCert)
	}()

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

func getLiteAPIServerStartJoinData(data phases.JoinData) (*clientcmdapi.Config, error) {
	klog.Info("%v", data.Cfg().Discovery.BootstrapToken)
	if data.Cfg().Discovery.BootstrapToken != nil {
		ipstr, _, err := net.SplitHostPort(data.Cfg().Discovery.BootstrapToken.APIServerEndpoint)
		if err == nil {
			klog.Info("%v", ipstr)
			masterIP = ipstr
		}
	}
	tlsBootstrapCfg, err := data.TLSBootstrapCfg()
	if err != nil {
		return nil, err
	}
	return tlsBootstrapCfg, nil
}

func initKubeClient(data phases.JoinData) (*kubernetes.Clientset, error) {
	tlsBootstrapCfg, err := getLiteAPIServerStartJoinData(data)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, cluster := range tlsBootstrapCfg.Clusters {
			cluster.Server = constant.LiteAPIServerAddr
		}
	}()
	// Write the bootstrap kubelet config file or the TLS-Bootstrapped kubelet config file down to disk
	klog.V(1).Infof("[kubelet-start] writing bootstrap kubelet config file at %s", kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	if err := kubeconfigutil.WriteToDisk(kubeadmconstants.GetBootstrapKubeletKubeConfigPath(), tlsBootstrapCfg); err != nil {
		return nil, errors.Wrap(err, "couldn't save bootstrap-kubelet.conf to disk")
	}
	bootstrapClient, err := kubeconfigutil.ClientSetFromFile(kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	if err != nil {
		return nil, errors.Errorf("couldn't create client from kubeconfig file %q", kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	}
	return bootstrapClient, nil
}

func deployLiteAPIServer(kubeClient *kubernetes.Clientset, nodeName string) error {
	liteApiServerConfigMap, err := kubeClient.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), constant.EDGE_CERT_CM, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := generateLiteAPIServerCert(liteApiServerConfigMap.Data); err != nil {
		klog.Errorf("Generate lite-apiserver cert, error: %v", err)
		return err
	}
	if err := createLiteAPIServerConfig(); err != nil {
		klog.Errorf("Create lite-apiserver config, error: %v", err)
		return err
	}

	if err := startLiteAPIServer(); err != nil {
		klog.Errorf("Start lite-apiserver, error: %v", err)
		return err
	}
	klog.Infof("Deploy lite-apiserver success!")
	return nil
}

func generateLiteAPIServerCert(liteApiServerConfigMap map[string]string) error {
	ca, ok := liteApiServerConfigMap[constant.KUBE_API_CA_CRT]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.KUBE_API_CA_CRT)
	}
	key, ok := liteApiServerConfigMap[constant.LITE_API_SERVER_KEY]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LITE_API_SERVER_KEY)
	}
	crt, ok := liteApiServerConfigMap[constant.LITE_API_SERVER_CRT]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LITE_API_SERVER_CRT)
	}
	tls, ok := liteApiServerConfigMap[constant.LITE_API_SERVER_TLS_CFG]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LITE_API_SERVER_TLS_CFG)
	}

	cmds := []string{
		fmt.Sprintf("mkdir -p %s", constant.KubeEdgePath),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteApiServerCACert, ca),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteApiserverKey, key),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteApiserverCrt, crt),
		fmt.Sprintf("cat << EOF >%s \n%s\nEOF", constant.LiteApiserverTLS, tls),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func createLiteAPIServerConfig() error {
	liteApiserverConfigTemplate := constant.LiteApiserverTemplate
	liteApiserverConfigTemplate = strings.ReplaceAll(liteApiserverConfigTemplate, "${MASTER_IP}", masterIP)
	cmds := []string{
		fmt.Sprintf(`echo "%s" > %s`, liteApiserverConfigTemplate, constant.LiteApiserverConfFile),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func startLiteAPIServer() error {
	cmds := []string{
		fmt.Sprintf(`cp %s %s`, workPath+constant.LiteApiserverBinPath, constant.UsrLocalBinDir),
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

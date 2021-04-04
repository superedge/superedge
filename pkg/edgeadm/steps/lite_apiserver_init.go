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
	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	phases "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/join"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	kubeconfigutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util/kubeconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
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
	kubeClient, err := initKubeClient(data)
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return err
	}
	// Deletes the bootstrapKubeConfigFile, so the credential used for TLS bootstrap is removed from disk
	defer os.Remove(kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	defer os.Remove(constant.KubeadmCert)

	// Deploy LiteAPIServer
	klog.Infof("Node: %s Start deploy LiteAPIServer", options.NodeName)
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

func getLiteAPIServerStartJoinData(data phases.JoinData) (*kubeadmapi.JoinConfiguration, *kubeadmapi.InitConfiguration, *clientcmdapi.Config, error) {
	cfg := data.Cfg()
	initCfg, err := data.InitCfg()
	if err != nil {
		return nil, nil, nil, err
	}
	klog.Info("%v", cfg.Discovery.BootstrapToken)
	if cfg.Discovery.BootstrapToken != nil {
		ipstr, _, err := net.SplitHostPort(cfg.Discovery.BootstrapToken.APIServerEndpoint)
		if err == nil {
			klog.Info("%v", ipstr)
			masterIP = ipstr
		}
	}
	tlsBootstrapCfg, err := data.TLSBootstrapCfg()
	if err != nil {
		return nil, nil, nil, err
	}
	return cfg, initCfg, tlsBootstrapCfg, nil
}

func initKubeClient(data phases.JoinData) (*kubernetes.Clientset, error) {
	_, _, tlsBootstrapCfg, err := getLiteAPIServerStartJoinData(data)
	if err != nil {
		return nil, err
	}

	// Write the ca certificate to disk so kubelet can use it for authentication
	cluster := tlsBootstrapCfg.Contexts[tlsBootstrapCfg.CurrentContext].Cluster
	if _, err := os.Stat(constant.LiteApiServerCACert); os.IsNotExist(err) {
		klog.V(1).Infof("[kubelet-start] writing CA certificate at %s", constant.LiteApiServerCACert)
		if err := certutil.WriteCert(constant.LiteApiServerCACert, tlsBootstrapCfg.Clusters[cluster].CertificateAuthorityData); err != nil {
			return nil, errors.Wrap(err, "couldn't save the CA certificate to disk")
		}
	}

	// Write the bootstrap kubelet config file or the TLS-Bootstrapped kubelet config file down to disk
	klog.V(1).Infof("[kubelet-start] writing bootstrap kubelet config file at %s", kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	if err := kubeconfigutil.WriteToDisk(kubeadmconstants.GetBootstrapKubeletKubeConfigPath(), tlsBootstrapCfg); err != nil {
		return nil, errors.Wrap(err, "couldn't save bootstrap-kubelet.conf to disk")
	}
	bootstrapClient, err := kubeconfigutil.ClientSetFromFile(kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	if err != nil {
		return nil, errors.Errorf("couldn't create client from kubeconfig file %q", kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
	}
	/*bootstrapClient, err = kubeclient.GetClientSet("")
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return nil, err
	}*/
	return bootstrapClient, nil
}

func deployLiteAPIServer(kubeClient *kubernetes.Clientset, nodeName string) error {
	liteApiServerConfigMap, err := kubeClient.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), constant.EDGE_CERT_CM, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if err := generateLiteAPIServerCert(kubeClient, liteApiServerConfigMap.Data); err != nil {
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

func generateLiteAPIServerCert(kubeClient *kubernetes.Clientset, liteApiServerConfigMap map[string]string) error {
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
		fmt.Sprintf("mkdir -p /etc/kubernetes/edge/"),
		fmt.Sprintf("cat << EOF >/etc/kubernetes/edge/lite-apiserver.key\n%s\nEOF", key),
		fmt.Sprintf("cat << EOF >/etc/kubernetes/edge/lite-apiserver.crt\n%s\nEOF", crt),
		fmt.Sprintf("cat << EOF >/etc/kubernetes/edge/tls.json \n%s\nEOF", tls),
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

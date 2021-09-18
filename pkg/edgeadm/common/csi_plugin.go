package common

import (
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests/topolvm"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"path/filepath"
)

func DeployTopolvmAppS(kubeconfigFile, manifestsDir, caCertFile, caKeyFile, masterPublicAddr string, certSANs []string) error {
	client, err := kubeclient.GetClientSet(kubeconfigFile)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}

	if err := CreateNamespace(client, constant.NamespaceTopolvmSystem); err != nil {
		return err
	}

	DeployTopolvmCRD(kubeconfigFile, manifestsDir)

	DeployTopolvmController(client, manifestsDir)

	DeployTopolvmScheduler(client, manifestsDir)

	DeployTopolvmWebhook(client, manifestsDir)

	DeployTopolvmLvmd(client, manifestsDir)

	DeployTopolvmNode(client, manifestsDir)

	klog.V(1).Infof("Deploy topolvm success!")

	return nil
}

func DeployTopolvmCRD(kubeconfigFile string, manifestsDir string) error {
	client, err := kubeclient.GetAPIExtensionSclientset(kubeconfigFile)
	if err != nil {
		return err
	}

	topolvmCRDConfig := filepath.Join(manifestsDir, manifests.AppTopolvmCRD)
	topolvmCRD := ReadYaml(topolvmCRDConfig, manifests.AppTopolvmCRDYaml)
	err = kubeclient.CreateOrUpdateCustomResourceDefinition(client, topolvmCRD, nil)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmCRD, err)
		return err
	}
	return err
}

// DeployTopolvmWebhook installs topolvm webhook to a Kubernetes cluster
func DeployTopolvmWebhook(client *kubernetes.Clientset, manifestsDir string) error {
	// Deploy topolvm-webhook config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmWebhookConfig := filepath.Join(manifestsDir, manifests.AppTopolvmWebhook)
	topolvmWebhook := ReadYaml(topolvmWebhookConfig, manifests.AppTopolvmWebhookYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmWebhook, option)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmWebhook, err)
		return err
	}

	return err
}

// DeployTopolvmController installs topolvm controller to a Kubernetes cluster
func DeployTopolvmController(client *kubernetes.Clientset, manifestsDir string) error {
	// Deploy topolvm-controller config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmControllerConfig := filepath.Join(manifestsDir, manifests.AppTopolvmController)
	topolvmController := ReadYaml(topolvmControllerConfig, manifests.AppTopolvmControllerYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmController, option)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmController, err)
		return err
	}

	return err
}

// DeployTopolvmscheduler installs topolvm scheduler to a Kubernetes cluster
func DeployTopolvmScheduler(client *kubernetes.Clientset, manifestsDir string) error {
	// Deploy topolvm-scheduler config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmWebhookConfig := filepath.Join(manifestsDir, manifests.AppTopolvmScheduler)
	topolvmWebhook := ReadYaml(topolvmWebhookConfig, manifests.AppTopolvmSchedulerYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmWebhook, option)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmScheduler, err)
		return err
	}

	return err
}

// DeployTopolvmLvmd installs topolvm-lvmd to a Kubernetes cluster
func DeployTopolvmLvmd(client *kubernetes.Clientset, manifestsDir string) error {
	// Deploy topolvm-lvmd config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmLvmdConfig := filepath.Join(manifestsDir, manifests.AppTopolvmLvmd)
	topolvmLvmd := ReadYaml(topolvmLvmdConfig, manifests.AppTopolvmLvmdYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmLvmd, option)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmLvmd, err)
		return err
	}

	return err
}

// DeployTopolvmNode installs topolvm-node to a Kubernetes cluster
func DeployTopolvmNode(client *kubernetes.Clientset, manifestsDir string) error {
	// Deploy topolvm-node config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmNodeConfig := filepath.Join(manifestsDir, manifests.AppTopolvmNode)
	topolvmNode := ReadYaml(topolvmNodeConfig, manifests.AppTopolvmNodeYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmNode, option)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmNode, err)
		return err
	}

	return err
}

func RemoveTopolvmApps(kubeconfigFile, manifestsDir, caCertFile, caKeyFile, masterPublicAddr string, certSANs []string) error {
	client, err := kubeclient.GetClientSet(kubeconfigFile)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}

	RemoveTopolvmNode(client, manifestsDir)

	RemoveTopolvmLvmd(client, manifestsDir)

	RemoveTopolvmScheduler(client, manifestsDir)

	RemoveTopolvmController(client, manifestsDir)

	RemoveTopolvmWebhook(client, manifestsDir)

	RemoveTopolvmCRD(client, manifestsDir)

	klog.V(1).Infof("Remove topolvm success!")

	return nil
}

// RemoveTopolvmLvmd remove topolvm-lvmd to a Kubernetes cluster
func RemoveTopolvmLvmd(client *kubernetes.Clientset, manifestsDir string) error {
	// Deploy topolvm-lvmd config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmLvmdConfig := filepath.Join(manifestsDir, manifests.AppTopolvmLvmd)
	topolvmLvmd := ReadYaml(topolvmLvmdConfig, manifests.AppTopolvmLvmdYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmLvmd, option)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmLvmd, err)
		return err
	}

	return err
}

// RemoveTopolvmNode remove topolvm-node to a Kubernetes cluster
func RemoveTopolvmNode(client *kubernetes.Clientset, manifestsDir string) error {
	// Remove topolvm-node config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmNodeConfig := filepath.Join(manifestsDir, manifests.AppTopolvmNode)
	topolvmNode := ReadYaml(topolvmNodeConfig, manifests.AppTopolvmNodeYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmNode, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmNode, err)
		return err
	}

	return err
}

// RemoveTopolvmscheduler remove topolvm scheduler to a Kubernetes cluster
func RemoveTopolvmScheduler(client *kubernetes.Clientset, manifestsDir string) error {
	// Remove topolvm-scheduler config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmWebhookConfig := filepath.Join(manifestsDir, manifests.AppTopolvmScheduler)
	topolvmWebhook := ReadYaml(topolvmWebhookConfig, manifests.AppTopolvmSchedulerYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmWebhook, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmScheduler, err)
		return err
	}

	return err
}

// RemoveTopolvmController remove topolvm controller to a Kubernetes cluster
func RemoveTopolvmController(client *kubernetes.Clientset, manifestsDir string) error {
	// Remove topolvm-controller config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmControllerConfig := filepath.Join(manifestsDir, manifests.AppTopolvmController)
	topolvmController := ReadYaml(topolvmControllerConfig, manifests.AppTopolvmControllerYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmController, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmController, err)
		return err
	}

	return err
}

// RemoveTopolvmWebhook remove topolvm webhook to a Kubernetes cluster
func RemoveTopolvmWebhook(client *kubernetes.Clientset, manifestsDir string) error {
	// Remove topolvm-webhook config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
	}
	topolvmWebhookConfig := filepath.Join(manifestsDir, manifests.AppTopolvmWebhook)
	topolvmWebhook := ReadYaml(topolvmWebhookConfig, manifests.AppTopolvmWebhookYaml)
	err := kubeclient.DeleteResourceWithFile(client, topolvmWebhook, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmWebhook, err)
		return err
	}

	return err
}

func RemoveTopolvmCRD(client *kubernetes.Clientset, manifestsDir string) error {
	topolvmCRDConfig := filepath.Join(manifestsDir, manifests.AppTopolvmCRD)
	topolvmCRD := ReadYaml(topolvmCRDConfig, manifests.AppTopolvmCRDYaml)
	err := kubeclient.CreateResourceWithFile(client, topolvmCRD, nil)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmCRD, err)
		return err
	}
	return err
}

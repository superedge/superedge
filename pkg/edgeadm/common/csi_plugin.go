package common

import (
	"encoding/base64"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests/topolvm"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"path/filepath"
)

func DeployTopolvmAppS(kubeconfigFile, manifestsDir string) error {
	client, err := kubeclient.GetClientSet(kubeconfigFile)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}

	if err := CreateNamespace(client, constant.NamespaceTopolvmSystem); err != nil {
		return err
	}

	if err := IgnoreTopolvmWebookNamesapceMatchExpressions(client); err != nil {
		klog.Errorf("Ignore topolvm-webook namesapce matchExpressions error: %v", err)
	}

	if err := DeployTopolvmCRD(kubeconfigFile, manifestsDir); err != nil {
		klog.Errorf("Deploy topolvm-crd error: %v", err)
		return err
	}

	if err := RemoveTopolvmWebhook(client, manifestsDir); err != nil {
		klog.V(4).Infof("Remove topolvm-webhook error: %v", err)
	}

	if err := DeployTopolvmWebhook(client, manifestsDir); err != nil {
		klog.Errorf("Deploy topolvm-webhook error: %v", err)
		return err
	}

	if err := DeployTopolvmController(client, manifestsDir); err != nil {
		klog.Errorf("Deploy topolvm-controller error: %v", err)
		return err
	}

	if err := DeployTopolvmScheduler(client, manifestsDir); err != nil {
		klog.Errorf("Deploy topolvm-scheduler error: %v", err)
		return err
	}

	if err := DeployTopolvmLvmd(client, manifestsDir); err != nil {
		klog.Errorf("Deploy topolvm-lvmd error: %v", err)
		return err
	}

	if err := DeployTopolvmNode(client, manifestsDir); err != nil {
		klog.Errorf("Deploy topolvm-node error: %v", err)
		return err
	}

	klog.V(1).Infof("Deploy topolvm all module success!")

	return nil
}

func IgnoreTopolvmWebookNamesapceMatchExpressions(client *kubernetes.Clientset) error {
	namespaceLabelMap := make(map[string]string)
	namespaceLabelMap[constant.TopolvmWebookMatchExpressionsKey] = constant.TopolvmWebookMatchExpressionsValue
	err := kubeclient.AddNameSpaceLabel(client, constant.NamespaceKubeSystem, namespaceLabelMap)
	err = kubeclient.AddNameSpaceLabel(client, constant.NamespaceTopolvmSystem, namespaceLabelMap)
	return err
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
	klog.V(1).Infof("Deploy %s success!", manifests.AppTopolvmCRD)
	return err
}

// DeployTopolvmWebhook installs topolvm webhook to a Kubernetes cluster
func DeployTopolvmWebhook(client *kubernetes.Clientset, manifestsDir string) error {
	caBundle, ca, caKey, err := util.GenerateCA(constant.OrganizationSuperEdgeIO)
	if err != nil {
		return err
	}

	dns := []string{
		"controller.topolvm-system.svc",
		"topolvm-controller.topolvm-system.svc",
		constant.OrganizationSuperEdgeIO,
	}
	serviceCert, serviceKey, err := util.GetServiceCertByRootca("topolvmWebhook", constant.OrganizationSuperEdgeIO, dns, ca, caKey)
	if err != nil {
		return err
	}

	// Deploy topolvm-webhook config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
		"CABundle":  caBundle,
		"ServerCrt": base64.StdEncoding.EncodeToString(serviceCert),
		"ServerKey": base64.StdEncoding.EncodeToString(serviceKey),
	}
	topolvmWebhookConfig := filepath.Join(manifestsDir, manifests.AppTopolvmWebhook)
	topolvmWebhook := ReadYaml(topolvmWebhookConfig, manifests.AppTopolvmWebhookYaml)
	err = kubeclient.CreateResourceWithFile(client, topolvmWebhook, option)
	if err != nil {
		klog.Errorf("Deploy %s config error: %v", manifests.AppTopolvmWebhook, err)
		return err
	}
	klog.V(1).Infof("Deploy %s success!", manifests.AppTopolvmWebhook)

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
	klog.V(1).Infof("Deploy %s success!", manifests.AppTopolvmController)

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
	klog.V(1).Infof("Deploy %s success!", manifests.AppTopolvmScheduler)

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
	klog.V(1).Infof("Deploy %s success!", manifests.AppTopolvmLvmd)

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
	klog.V(1).Infof("Deploy %s success!", manifests.AppTopolvmNode)

	return err
}

func RemoveTopolvmApps(kubeconfigFile, manifestsDir string) error {
	client, err := kubeclient.GetClientSet(kubeconfigFile)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return err
	}

	if err := RemoveTopolvmNode(client, manifestsDir); err != nil {
		klog.Errorf("Remove topolvm-node from your cluster, error: %v", err)
		return err
	}

	if err := RemoveTopolvmLvmd(client, manifestsDir); err != nil {
		klog.Errorf("Remove topolvm-lvm from your cluster, error: %v", err)
		return err
	}

	if err := RemoveTopolvmScheduler(client, manifestsDir); err != nil {
		klog.Errorf("Remove topolvm-scheduler from your cluster, error: %v", err)
		return err
	}

	if err := RemoveTopolvmController(client, manifestsDir); err != nil {
		klog.Errorf("Remove topolvm-controller from your cluster, error: %v", err)
		return err
	}

	if err := RemoveTopolvmWebhook(client, manifestsDir); err != nil {
		klog.Errorf("Remove topolvm-webhook from your cluster, error: %v", err)
		return err
	}

	//RemoveTopolvmCRD(client, manifestsDir)

	klog.V(1).Infof("Remove topolvm all module success!")

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
	err := kubeclient.DeleteResourceWithFile(client, topolvmLvmd, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmLvmd, err)
		return err
	}
	klog.V(1).Infof("Remove %s success!", manifests.AppTopolvmLvmd)

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
	err := kubeclient.DeleteResourceWithFile(client, topolvmNode, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmNode, err)
		return err
	}
	klog.V(1).Infof("Remove %s success!", manifests.AppTopolvmNode)

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
	err := kubeclient.DeleteResourceWithFile(client, topolvmWebhook, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmScheduler, err)
		return err
	}
	klog.V(1).Infof("Remove %s success!", manifests.AppTopolvmScheduler)

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
	err := kubeclient.DeleteResourceWithFile(client, topolvmController, option)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmController, err)
		return err
	}
	klog.V(1).Infof("Remove %s success!", manifests.AppTopolvmController)

	return err
}

// RemoveTopolvmWebhook remove topolvm webhook to a Kubernetes cluster
func RemoveTopolvmWebhook(client *kubernetes.Clientset, manifestsDir string) error {
	// Remove topolvm-webhook config
	option := map[string]interface{}{
		"Namespace": constant.NamespaceTopolvmSystem,
		"CABundle":  "",
		"ServerCrt": "",
		"ServerKey": "",
	}
	topolvmWebhookConfig := filepath.Join(manifestsDir, manifests.AppTopolvmWebhook)
	topolvmWebhook := ReadYaml(topolvmWebhookConfig, manifests.AppTopolvmWebhookYaml)
	err := kubeclient.DeleteResourceWithFile(client, topolvmWebhook, option)
	if err != nil {
		return err
	}
	klog.V(1).Infof("Remove %s success!", manifests.AppTopolvmWebhook)

	return err
}

func RemoveTopolvmCRD(client *kubernetes.Clientset, manifestsDir string) error {
	topolvmCRDConfig := filepath.Join(manifestsDir, manifests.AppTopolvmCRD)
	topolvmCRD := ReadYaml(topolvmCRDConfig, manifests.AppTopolvmCRDYaml)
	err := kubeclient.DeleteResourceWithFile(client, topolvmCRD, nil)
	if err != nil {
		klog.Errorf("Remove %s config error: %v", manifests.AppTopolvmCRD, err)
		return err
	}
	return err
}

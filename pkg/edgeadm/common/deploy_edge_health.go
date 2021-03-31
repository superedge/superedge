package common

import (
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"path/filepath"
)

func DeployEdgeHealth(clientSet kubernetes.Interface, manifestsDir string) error {
	userEdgeHealthWebhook := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH_WEBHOOK)
	userEdgeHealthAdmission := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH_ADMISSION)
	yamlMap := map[string]string{
		manifests.APP_EDGE_HEALTH_ADMISSION: ReadYaml(userEdgeHealthAdmission, manifests.EdgeHealthAdmissionYaml),
		manifests.APP_EDGE_HEALTH_WEBHOOK:   ReadYaml(userEdgeHealthWebhook, manifests.EdgeHealthWebhookConfigYaml),
	}
	for appName, yamlFile := range yamlMap {
		if err := CreateByYamlFile(clientSet, yamlFile); err != nil {
			return err
		}
		klog.Infof("Create %s success!\n", appName)
	}

	option := map[string]interface{}{
		"HmacKey": util.GetRandToken(16),
	}

	userManifests := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH)
	edgeHealthYaml := ReadYaml(userManifests, manifests.EdgeHealthYaml)
	if err := kubeclient.CreateResourceWithFile(clientSet, edgeHealthYaml, option); err != nil {
		return err
	}
	klog.Infof("Create %s success!\n", manifests.APP_EDGE_HEALTH)

	return nil
}

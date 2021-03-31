package common

import (
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"path/filepath"
)

func DeploySerivceGroup(clientSet kubernetes.Interface, manifestsDir string) error {
	userGridWrapper := filepath.Join(manifestsDir, manifests.APP_APPLICATION_GRID_WRAPPER)
	userGridController := filepath.Join(manifestsDir, manifests.APP_APPLICATION_GRID_CONTROLLER)
	yamlMap := map[string]string{
		manifests.APP_EDGE_HEALTH_ADMISSION: ReadYaml(userGridWrapper, manifests.ApplicationGridWrapperYaml),
		manifests.APP_EDGE_HEALTH_WEBHOOK:   ReadYaml(userGridController, manifests.ApplicationGridControllerYaml),
	}

	for appName, yamlFile := range yamlMap {
		if err := CreateByYamlFile(clientSet, yamlFile); err != nil {
			return err
		}
		klog.Infof("Create %s success!\n", appName)
	}
	klog.Infof("Create %s success!\n", manifests.APP_EDGE_HEALTH)

	return nil
}

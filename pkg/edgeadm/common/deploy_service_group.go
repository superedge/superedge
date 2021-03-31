package common

import (
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"path/filepath"
)

func DeployServiceGroup(clientSet kubernetes.Interface, manifestsDir, advertiseAddress string, bindPort int32) error {
	option := map[string]interface{}{
		"BindPort":         bindPort,
		"AdvertiseAddress": advertiseAddress,
	}
	userGridWrapper := filepath.Join(manifestsDir, manifests.APP_APPLICATION_GRID_WRAPPER)
	gridWrapper := ReadYaml(userGridWrapper, manifests.ApplicationGridWrapperYaml)
	if err := kubeclient.CreateResourceWithFile(clientSet, gridWrapper, option); err != nil {
		return err
	}
	klog.V(4).Infof("Deploy %s success!", manifests.APP_APPLICATION_GRID_WRAPPER)

	userGridController := filepath.Join(manifestsDir, manifests.APP_APPLICATION_GRID_CONTROLLER)
	gridController := ReadYaml(userGridController, manifests.ApplicationGridControllerYaml)
	if err := CreateByYamlFile(clientSet, gridController); err != nil {
		klog.Errorf("Deploy %s error: %s", manifests.APP_APPLICATION_GRID_CONTROLLER, err)
		return err
	}

	klog.V(4).Infof("Create %s success!", manifests.APP_APPLICATION_GRID_CONTROLLER)

	return nil
}

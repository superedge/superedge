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

package common

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"path/filepath"

	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func DeployEdgeHealth(clientSet kubernetes.Interface, manifestsDir string) error {
	yamlMap, edgeHealthYaml, option, err := getEdgeHealthResource(clientSet, manifestsDir)
	if err != nil {
		return err
	}
	for appName, yamlFile := range yamlMap {
		if err := CreateByYamlFile(clientSet, yamlFile); err != nil {
			return err
		}
		klog.Infof("Create %s success!\n", appName)
	}
	if err := kubeclient.CreateResourceWithFile(clientSet, edgeHealthYaml, option); err != nil {
		return err
	}
	klog.Infof("Create %s success!\n", manifests.APP_EDGE_HEALTH)

	return nil
}

func DeleteEdgeHealth(clientSet kubernetes.Interface, manifestsDir string) error {
	yamlMap, edgeHealthYaml, option, err := getEdgeHealthResource(clientSet, manifestsDir)
	if err != nil {
		return err
	}
	for appName, yamlFile := range yamlMap {
		if err := DeleteByYamlFile(clientSet, yamlFile); err != nil {
			return err
		}
		klog.Infof("Delete %s success!\n", appName)
	}
	if err := kubeclient.DeleteResourceWithFile(clientSet, edgeHealthYaml, option); err != nil {
		return err
	}
	klog.Infof("Delete %s success!\n", manifests.APP_EDGE_HEALTH)

	return nil
}

func getEdgeHealthResource(clientSet kubernetes.Interface, manifestsDir string) (map[string]string, string, interface{}, error) {
	userEdgeHealthWebhook := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH_WEBHOOK)
	userEdgeHealthAdmission := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH_ADMISSION)
	yamlMap := map[string]string{
		manifests.APP_EDGE_HEALTH_ADMISSION: ReadYaml(userEdgeHealthAdmission, manifests.EdgeHealthAdmissionYaml),
		manifests.APP_EDGE_HEALTH_WEBHOOK:   ReadYaml(userEdgeHealthWebhook, manifests.EdgeHealthWebhookConfigYaml),
	}

	option := map[string]interface{}{
		"HmacKey": util.GetRandToken(16),
	}

	userManifests := filepath.Join(manifestsDir, manifests.APP_EDGE_HEALTH)
	edgeHealthYaml := ReadYaml(userManifests, manifests.EdgeHealthYaml)
	return yamlMap, edgeHealthYaml, option, nil
}

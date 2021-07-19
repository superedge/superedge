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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
)

var yamlMap = map[string]string{
	manifests.KUBE_FLANNEL:                    manifests.KubeFlannelYaml,
	manifests.APP_HELPER_JOB:                  manifests.HelperJobYaml,
	manifests.APP_EDGE_HEALTH:                 manifests.EdgeHealthYaml,
	manifests.APP_TUNNEL_EDGE:                 manifests.TunnelEdgeYaml,
	manifests.APP_TUNNEL_CLOUD:                manifests.TunnelCloudYaml,
	manifests.APP_TUNNEL_CORDDNS:              manifests.TunnelCorednsYaml,
	manifests.APP_LITE_APISERVER:              manifests.LiteApiServerYaml,
	manifests.APPEdgeCorednsConfig:            manifests.EdgeCorednsConfigYaml,
	manifests.APP_EDGE_HEALTH_ADMISSION:       manifests.EdgeHealthAdmissionYaml,
	manifests.APP_EDGE_HEALTH_WEBHOOK:         manifests.EdgeHealthWebhookConfigYaml,
	manifests.APP_APPLICATION_GRID_WRAPPER:    manifests.ApplicationGridWrapperYaml,
	manifests.APPEdgeCorednsServiceGrid:       manifests.EdgeCorednsServiceGridYaml,
	manifests.APPEdgeCorednsDeploymentGrid:    manifests.EdgeCorednsDeploymentGridYaml,
	manifests.APP_APPLICATION_GRID_CONTROLLER: manifests.ApplicationGridControllerYaml,
	manifests.KUBE_VIP:                        manifests.KubeVIPYaml,
}

func NewManifestsCMD() *cobra.Command {
	var yamlDir string
	cmd := &cobra.Command{
		Use:   "manifests",
		Short: "Output edge cluster manifest yaml files",
		Run: func(cmd *cobra.Command, args []string) {
			if err := outputYamlFile(yamlDir); err != nil {
				util.OutPutMessage(err.Error())
				return
			}
		},
	}

	cmd.Flags().StringVarP(&yamlDir, "manifest-dir", "m", "./manifests/",
		"Folder for edge cluster yaml files output.")

	return cmd
}

func outputYamlFile(yamlPath string) error {
	if !util.IsFileExist(yamlPath) {
		if err := os.MkdirAll(yamlPath, os.ModePerm); err != nil {
			return err
		}
	}

	for yamlName := range yamlMap {
		filePath := yamlPath + "/" + yamlName
		if !util.IsFileExist(filePath) {
			util.WriteWithBufio(filePath, yamlMap[yamlName])
		}
	}

	fmt.Printf("Success output edge clsuter yaml files to %s\n", yamlPath)
	return nil
}

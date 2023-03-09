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

package util

import (
	"fmt"

	crdv1 "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io/v1"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"github.com/superedge/superedge/pkg/application-grid-controller/util"
	commonutil "github.com/superedge/superedge/pkg/application-grid-controller/util"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

type DeploymentTemplateHash struct{}

func NewDeploymentTemplateHash() DeploymentTemplateHash {
	return DeploymentTemplateHash{}
}

func (dth *DeploymentTemplateHash) RemoveUnusedTemplate(dg *crdv1.DeploymentGrid) error {
	if !dg.Spec.AutoDeleteUnusedTemplate {
		return nil
	}
	templateMap := make(map[string]bool, len(dg.Spec.TemplatePool))
	for k := range dg.Spec.TemplatePool {
		templateMap[k] = true
	}
	if dg.Spec.DefaultTemplateName != common.DefaultTemplateName {
		delete(templateMap, dg.Spec.DefaultTemplateName)
	}
	for _, template_used := range dg.Spec.Templates {
		delete(templateMap, template_used)
	}
	for template_unused := range templateMap {
		delete(dg.Spec.TemplatePool, template_unused)
	}
	return nil
}

func (dth *DeploymentTemplateHash) UpdateTemplateHash(dg *crdv1.DeploymentGrid) {
	updateHash := func(template *appsv1.DeploymentSpec) {
		dth.setTemplateHash(template)
	}

	updateHash(&dg.Spec.Template)

	for _, template := range dg.Spec.TemplatePool {
		updateHash(&template)
	}
}

func (dth *DeploymentTemplateHash) setTemplateHash(template *appsv1.DeploymentSpec) {
	expected := dth.generateTemplateHash(template)
	hash := util.GetTemplateHash(template.Template.Labels)
	if hash != expected {
		if template.Template.Labels == nil {
			template.Template.Labels = make(map[string]string)
		}
		template.Template.Labels[common.TemplateHashKey] = expected
	}
}

func (dth *DeploymentTemplateHash) generateTemplateHash(template *appsv1.DeploymentSpec) string {
	meta := template.Template.ObjectMeta.DeepCopy()
	copyTemplate := template.DeepCopy()
	delete(meta.Labels, common.TemplateHashKey)
	copyTemplate.Template.ObjectMeta = *meta
	// replicas doesn't need hash caculation
	copyTemplate.Replicas = nil
	return fmt.Sprintf("%d", util.GenerateHash(copyTemplate))
}

func (dth *DeploymentTemplateHash) IsTemplateHashChanged(dg *crdv1.DeploymentGrid, gridValues string, dp *appsv1.Deployment) bool {
	hash := util.GetTemplateHash(dp.Spec.Template.Labels)

	template, err := dth.getDeployTemplate(&dg.Spec, gridValues)
	if err != nil {
		klog.Errorf("Failed to get deployment template for %s from deploymentGrid %s", dp.Name, dg.Name)
		return true
	}
	expected := util.GetTemplateHash(template.Template.Labels)
	return hash != expected
}

func (dth *DeploymentTemplateHash) getDeployTemplate(spec *crdv1.DeploymentGridSpec, gridValues string) (*appsv1.DeploymentSpec, error) {
	templateName := commonutil.GetTemplateName(spec.Templates, gridValues, spec.DefaultTemplateName)
	if templateName == common.DefaultTemplateName || templateName == "" {
		return &spec.Template, nil
	} else if template, ok := spec.TemplatePool[templateName]; ok {
		return &template, nil
	} else {
		return nil, fmt.Errorf("template not found in templatePool")
	}
}
func (dth *DeploymentTemplateHash) IsReplicasChanged(dg *crdv1.DeploymentGrid, gridValues string, dp *appsv1.Deployment) bool {
	template, err := dth.getDeployTemplate(&dg.Spec, gridValues)
	if err != nil {
		klog.Errorf("Failed to get deployment template for %s from deploymentGrid %s", dp.Name, dg.Name)
		return true
	}

	return !(*template.Replicas == *dp.Spec.Replicas)
}

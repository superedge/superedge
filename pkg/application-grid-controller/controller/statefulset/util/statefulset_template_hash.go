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

type StatefulsetTemplateHash struct{}

func NewStatefulsetTemplateHash() StatefulsetTemplateHash {
	return StatefulsetTemplateHash{}
}

func (sth *StatefulsetTemplateHash) RemoveUnusedTemplate(ssg *crdv1.StatefulSetGrid) error {
	if !ssg.Spec.AutoDeleteUnusedTemplate {
		return nil
	}
	templateMap := make(map[string]bool, len(ssg.Spec.TemplatePool))
	for k := range ssg.Spec.TemplatePool {
		templateMap[k] = true
	}
	if ssg.Spec.DefaultTemplateName != common.DefaultTemplateName {
		delete(templateMap, ssg.Spec.DefaultTemplateName)
	}
	for _, template_used := range ssg.Spec.Templates {
		delete(templateMap, template_used)
	}
	for template_unused := range templateMap {
		delete(ssg.Spec.TemplatePool, template_unused)
	}
	return nil
}

func (sth *StatefulsetTemplateHash) UpdateTemplateHash(ssg *crdv1.StatefulSetGrid) {
	updateHash := func(template *appsv1.StatefulSetSpec) {
		sth.setTemplateHash(template)
	}

	updateHash(&ssg.Spec.Template)

	for _, template := range ssg.Spec.TemplatePool {
		updateHash(&template)
	}
}

func (sth *StatefulsetTemplateHash) setTemplateHash(template *appsv1.StatefulSetSpec) {
	expected := sth.generateTemplateHash(template)
	hash := util.GetTemplateHash(template.Template.Labels)
	if hash != expected {
		if template.Template.Labels == nil {
			template.Template.Labels = make(map[string]string)
		}
		template.Template.Labels[common.TemplateHashKey] = expected
	}
}

func (sth *StatefulsetTemplateHash) generateTemplateHash(template *appsv1.StatefulSetSpec) string {
	meta := template.Template.ObjectMeta.DeepCopy()
	copyTemplate := template.DeepCopy()
	delete(meta.Labels, common.TemplateHashKey)
	copyTemplate.Template.ObjectMeta = *meta
	// replicas doesn't need hash caculation
	copyTemplate.Replicas = nil
	return fmt.Sprintf("%d", util.GenerateHash(copyTemplate))
}

func (sth *StatefulsetTemplateHash) IsTemplateHashChanged(ssg *crdv1.StatefulSetGrid, gridValues string, ss *appsv1.StatefulSet) bool {
	hash := util.GetTemplateHash(ss.Spec.Template.Labels)

	template, err := sth.getStatefulsetTemplate(&ssg.Spec, gridValues)
	if err != nil {
		klog.Errorf("Failed to get statefulset template for %s from statefulsetGrid %s", ss.Name, ssg.Name)
		return true
	}
	expected := util.GetTemplateHash(template.Template.Labels)
	return hash != expected
}

func (sth *StatefulsetTemplateHash) getStatefulsetTemplate(spec *crdv1.StatefulSetGridSpec, gridValues string) (*appsv1.StatefulSetSpec, error) {
	templateName := commonutil.GetTemplateName(spec.Templates, gridValues, spec.DefaultTemplateName)
	if templateName == common.DefaultTemplateName || templateName == "" {
		return &spec.Template, nil
	} else if template, ok := spec.TemplatePool[templateName]; ok {
		return &template, nil
	} else {
		return nil, fmt.Errorf("template not found in templatePool")
	}
}

func (sth *StatefulsetTemplateHash) IsReplicasChanged(ssg *crdv1.StatefulSetGrid, gridValues string, ss *appsv1.StatefulSet) bool {
	template, err := sth.getStatefulsetTemplate(&ssg.Spec, gridValues)
	if err != nil {
		klog.Errorf("Failed to get statefulset template for %s from statefulsetGrid %s", ss.Name, ssg.Name)
		return true
	}

	return !(*template.Replicas == *ss.Spec.Replicas)
}

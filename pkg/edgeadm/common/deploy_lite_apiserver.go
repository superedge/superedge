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
	"context"
	"errors"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/edgeadm/constant/manifests"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
)

func CreateLiteApiServerCert(clientSet kubernetes.Interface, manifestsDir, caCertFile, caKeyFile string) error {
	if err := kubeclient.CreateOrUpdateNamespace(clientSet, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.NamespaceEdgeSystem,
		},
	}); err != nil {
		return err
	}

	role := rbacv1.Role{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lite-apiserver",
			Namespace: constant.NamespaceEdgeSystem,
		},
		Rules: nil,
	}
	role.Rules = append(role.Rules, rbacv1.PolicyRule{
		APIGroups:     []string{"*"},
		Resources:     []string{"configmaps"},
		ResourceNames: []string{constant.EdgeCertCM},
		Verbs:         []string{"get", "list", "watch"},
	})
	roleBinding := rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lite-apiserver",
			Namespace: constant.NamespaceEdgeSystem,
		},
		RoleRef: rbacv1.RoleRef{
			Name:     "lite-apiserver",
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: nil,
	}
	roleBinding.Subjects = append(roleBinding.Subjects, rbacv1.Subject{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Group",
		Name:     "system:bootstrappers:kubeadm:default-node-token",
	})

	if _, err := clientSet.RbacV1().Roles(constant.NamespaceEdgeSystem).Create(
		context.TODO(), &role, metav1.CreateOptions{}); err != nil {
		return err
	}
	if _, err := clientSet.RbacV1().RoleBindings(constant.NamespaceEdgeSystem).Create(
		context.TODO(), &roleBinding, metav1.CreateOptions{}); err != nil {
		return err
	}

	clientSet.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).Delete(
		context.TODO(), constant.EdgeCertCM, metav1.DeleteOptions{})

	kubeService, err := clientSet.CoreV1().Services(
		constant.NamespaceDefault).Get(context.TODO(), constant.ServiceKubernetes, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if kubeService.Spec.ClusterIP == "" {
		return errors.New("Get kubernetes service clusterIP nil\n")
	}
	kubeAPIClusterIP := kubeService.Spec.ClusterIP

	liteApiServerCrt, liteApiServerKey, err :=
		GetServiceCert("LiteApiServer", caCertFile, caKeyFile, []string{"127.0.0.1"}, []string{kubeAPIClusterIP})
	if err != nil {
		return err
	}

	caCertStr, err := util.ReadFile(caCertFile)
	if err != nil {
		return err
	}
	userLiteAPIServer := filepath.Join(manifestsDir, manifests.APP_lITE_APISERVER)
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: constant.EdgeCertCM,
		},
		Data: map[string]string{
			constant.KubeAPICACrt:         string(caCertStr),
			constant.KubeAPIClusterIP:     kubeAPIClusterIP,
			constant.LiteAPIServerCrt:     string(liteApiServerCrt),
			constant.LiteAPIServerKey:     string(liteApiServerKey),
			constant.LiteAPIServerTLSJSON: constant.LiteAPIServerTLSCfg,
			manifests.APP_lITE_APISERVER:  ReadYaml(userLiteAPIServer, manifests.LiteApiServerYaml),
		},
	}

	if _, err := clientSet.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).
		Create(context.TODO(), configMap, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func DeleteLiteApiServerCert(clientSet kubernetes.Interface) error {
	if err := clientSet.RbacV1().Roles(constant.NamespaceEdgeSystem).Delete(
		context.TODO(), "lite-apiserver", metav1.DeleteOptions{}); err != nil {
		return err
	}
	if err := clientSet.RbacV1().RoleBindings(constant.NamespaceEdgeSystem).Delete(
		context.TODO(), "lite-apiserver", metav1.DeleteOptions{}); err != nil {
		return err
	}

	clientSet.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).Delete(
		context.TODO(), constant.EdgeCertCM, metav1.DeleteOptions{})

	return nil
}

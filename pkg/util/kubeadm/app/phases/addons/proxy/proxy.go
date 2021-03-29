/*
Copyright 2017 The Kubernetes Authors.

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

package proxy

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/componentconfigs"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/images"
	kubeadmutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/util/apiclient"
)

const (
	// KubeProxyClusterRoleName sets the name for the kube-proxy ClusterRole
	// TODO: This k8s-generic, well-known constant should be fetchable from another source, not be in this package
	KubeProxyClusterRoleName = "system:node-proxier"

	// KubeProxyServiceAccountName describes the name of the ServiceAccount for the kube-proxy addon
	KubeProxyServiceAccountName = "kube-proxy"
)

// EnsureProxyAddon creates the kube-proxy addons
func EnsureProxyAddon(cfg *kubeadmapi.ClusterConfiguration, localEndpoint *kubeadmapi.APIEndpoint, client clientset.Interface) error {
	if err := CreateServiceAccount(client); err != nil {
		return errors.Wrap(err, "error when creating kube-proxy service account")
	}

	if err := createKubeProxyConfigMap(cfg, localEndpoint, client); err != nil {
		return err
	}

	if err := createKubeProxyAddon(cfg, client); err != nil {
		return err
	}

	if err := CreateRBACRules(client); err != nil {
		return errors.Wrap(err, "error when creating kube-proxy RBAC rules")
	}

	fmt.Println("[addons] Applied essential addon: kube-proxy")
	return nil
}

// CreateServiceAccount creates the necessary serviceaccounts that kubeadm uses/might use, if they don't already exist.
func CreateServiceAccount(client clientset.Interface) error {

	return apiclient.CreateOrUpdateServiceAccount(client, &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeProxyServiceAccountName,
			Namespace: metav1.NamespaceSystem,
		},
	})
}

// CreateRBACRules creates the essential RBAC rules for a minimally set-up cluster
func CreateRBACRules(client clientset.Interface) error {
	return createClusterRoleBindings(client)
}

func createKubeProxyConfigMap(cfg *kubeadmapi.ClusterConfiguration, localEndpoint *kubeadmapi.APIEndpoint, client clientset.Interface) error {
	// Generate ControlPlane Enpoint kubeconfig file
	controlPlaneEndpoint, err := kubeadmutil.GetControlPlaneEndpoint(cfg.ControlPlaneEndpoint, localEndpoint)
	if err != nil {
		return err
	}

	kubeProxyCfg, ok := cfg.ComponentConfigs[componentconfigs.KubeProxyGroup]
	if !ok {
		return errors.New("no kube-proxy component config found in the active component config set")
	}

	proxyBytes, err := kubeProxyCfg.Marshal()
	if err != nil {
		return errors.Wrap(err, "error when marshaling")
	}
	var prefixBytes bytes.Buffer
	apiclient.PrintBytesWithLinePrefix(&prefixBytes, proxyBytes, "    ")
	configMapBytes, err := kubeadmutil.ParseTemplate(KubeProxyConfigMap19,
		struct {
			ControlPlaneEndpoint string
			ProxyConfig          string
			ProxyConfigMap       string
			ProxyConfigMapKey    string
		}{
			ControlPlaneEndpoint: controlPlaneEndpoint,
			ProxyConfig:          prefixBytes.String(),
			ProxyConfigMap:       constants.KubeProxyConfigMap,
			ProxyConfigMapKey:    constants.KubeProxyConfigMapKey,
		})
	if err != nil {
		return errors.Wrap(err, "error when parsing kube-proxy configmap template")
	}

	kubeproxyConfigMap := &v1.ConfigMap{}
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), configMapBytes, kubeproxyConfigMap); err != nil {
		return errors.Wrap(err, "unable to decode kube-proxy configmap")
	}

	if !kubeProxyCfg.IsUserSupplied() {
		componentconfigs.SignConfigMap(kubeproxyConfigMap)
	}

	// Create the ConfigMap for kube-proxy or update it in case it already exists
	return apiclient.CreateOrUpdateConfigMap(client, kubeproxyConfigMap)
}

func createKubeProxyAddon(cfg *kubeadmapi.ClusterConfiguration, client clientset.Interface) error {
	daemonSetbytes, err := kubeadmutil.ParseTemplate(KubeProxyDaemonSet19, struct{ Image, ProxyConfigMap, ProxyConfigMapKey string }{
		Image:             images.GetKubernetesImage(constants.KubeProxy, cfg),
		ProxyConfigMap:    constants.KubeProxyConfigMap,
		ProxyConfigMapKey: constants.KubeProxyConfigMapKey,
	})
	if err != nil {
		return errors.Wrap(err, "error when parsing kube-proxy daemonset template")
	}

	kubeproxyDaemonSet := &apps.DaemonSet{}
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), daemonSetbytes, kubeproxyDaemonSet); err != nil {
		return errors.Wrap(err, "unable to decode kube-proxy daemonset")
	}
	// Propagate the http/https proxy host environment variables to the container
	env := &kubeproxyDaemonSet.Spec.Template.Spec.Containers[0].Env
	*env = append(*env, kubeadmutil.GetProxyEnvVars()...)

	// Create the DaemonSet for kube-proxy or update it in case it already exists
	return apiclient.CreateOrUpdateDaemonSet(client, kubeproxyDaemonSet)
}

func createClusterRoleBindings(client clientset.Interface) error {
	if err := apiclient.CreateOrUpdateClusterRoleBinding(client, &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeadm:node-proxier",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     KubeProxyClusterRoleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      KubeProxyServiceAccountName,
				Namespace: metav1.NamespaceSystem,
			},
		},
	}); err != nil {
		return err
	}

	// Create a role for granting read only access to the kube-proxy component config ConfigMap
	if err := apiclient.CreateOrUpdateRole(client, &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KubeProxyConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Rules: []rbac.PolicyRule{
			{
				Verbs:         []string{"get"},
				APIGroups:     []string{""},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{constants.KubeProxyConfigMap},
			},
		},
	}); err != nil {
		return err
	}

	// Bind the role to bootstrap tokens for allowing fetchConfiguration during join
	return apiclient.CreateOrUpdateRoleBinding(client, &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KubeProxyConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "Role",
			Name:     constants.KubeProxyConfigMap,
		},
		Subjects: []rbac.Subject{
			{
				Kind: rbac.GroupKind,
				Name: constants.NodeBootstrapTokenAuthGroup,
			},
		},
	})
}

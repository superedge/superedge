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

package kubeclient

import (
	"context"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	apps "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func DeleteConfigMap(client clientset.Interface, cm *corev1.ConfigMap) error {
	return client.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Delete(context.TODO(), cm.Name, metav1.DeleteOptions{})
}

func DeleteSecret(client clientset.Interface, secret *corev1.Secret) error {
	return client.CoreV1().Secrets(secret.ObjectMeta.Namespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{})
}

func DeleteService(client clientset.Interface, svc *corev1.Service) error {
	return client.CoreV1().Services(svc.ObjectMeta.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{})
}

func DeleteServiceAccount(client clientset.Interface, sa *corev1.ServiceAccount) error {
	return client.CoreV1().ServiceAccounts(sa.ObjectMeta.Namespace).Delete(context.TODO(), sa.Name, metav1.DeleteOptions{})
}

func DeleteDaemonSet(client clientset.Interface, ds *apps.DaemonSet) error {
	return client.AppsV1().DaemonSets(ds.ObjectMeta.Namespace).Delete(context.TODO(), ds.Name, metav1.DeleteOptions{})
}

func DeleteDeployment(client clientset.Interface, namespace string, deployName string, isExtensions bool, labelSelector metav1.LabelSelector) error {
	if isExtensions {
		return deleteDeploymentExtensionsV1beta1(client, deployName, labelSelector)
	} else {
		return deleteDeploymentAppsV1(client, namespace, deployName)
	}
}

func DeleteRole(client clientset.Interface, role *rbac.Role) error {
	return client.RbacV1().Roles(role.ObjectMeta.Namespace).Delete(context.TODO(), role.Name, metav1.DeleteOptions{})
}

func DeleteRoleBinding(client clientset.Interface, roleBinding *rbac.RoleBinding) error {
	return client.RbacV1().RoleBindings(roleBinding.ObjectMeta.Namespace).Delete(context.TODO(), roleBinding.Name, metav1.DeleteOptions{})
}

func DeleteClusterRole(client clientset.Interface, clusterRole *rbac.ClusterRole) error {
	return client.RbacV1().ClusterRoles().Delete(context.TODO(), clusterRole.Name, metav1.DeleteOptions{})
}

func DeleteClusterRoleBinding(client clientset.Interface, clusterRoleBinding *rbac.ClusterRoleBinding) error {
	return client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBinding.Name, metav1.DeleteOptions{})
}

func DeleteCSIDriver(client clientset.Interface, csiDriver *storagev1.CSIDriver) error {
	return client.StorageV1().CSIDrivers().Delete(context.TODO(), csiDriver.Name, metav1.DeleteOptions{})
}

func DeleteStorageClasses(client clientset.Interface, csiDriver *storagev1.StorageClass) error {
	return client.StorageV1().StorageClasses().Delete(context.TODO(), csiDriver.Name, metav1.DeleteOptions{})
}

func DeletePodSecurityPolicy(client clientset.Interface, podSecurityPolicy *v1beta1.PodSecurityPolicy) error {
	return client.PolicyV1beta1().PodSecurityPolicies().Delete(context.TODO(), podSecurityPolicy.Name, metav1.DeleteOptions{})
}

func DeleteMutatingMutatingWebhookConfigurations(client clientset.Interface, obj *admissionv1.MutatingWebhookConfiguration) error {
	return client.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
}

func DeleteJob(client clientset.Interface, job *batchv1.Job) error {
	return client.BatchV1().Jobs(job.ObjectMeta.Namespace).Delete(context.TODO(), job.Name, metav1.DeleteOptions{})
}

func IsJobFinished(j *batchv1.Job) bool {
	for _, c := range j.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

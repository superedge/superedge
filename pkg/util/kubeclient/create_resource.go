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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	apps "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	// APICallRetryInterval defines how long should wait before retrying a failed API operation
	APICallRetryInterval = 500 * time.Millisecond
	// PatchNodeTimeout specifies how long should wait for applying the label and taint on the master before timing out
	PatchNodeTimeout = 2 * time.Minute
	// UpdateNodeTimeout specifies how long should wait for updating node with the initial remote configuration of kubelet before timing out
	UpdateNodeTimeout = 2 * time.Minute
	// LabelHostname specifies the lable in node.
	LabelHostname = "kubernetes.io/hostname"
)

// CreateOrUpdateConfigMap creates a ConfigMap if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateConfigMap(client clientset.Interface, cm *corev1.ConfigMap) error {
	if _, err := client.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create configmap")
		}

		if _, err := client.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update configmap")
		}
	}
	return nil
}

// CreateOrRetainConfigMap creates a ConfigMap if the target resource doesn't exist. If the resource exists already, this function will retain the resource instead.
func CreateOrRetainConfigMap(client clientset.Interface, cm *corev1.ConfigMap, configMapName string) error {
	if _, err := client.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Get(context.TODO(), configMapName, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil
		}
		if _, err := client.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{}); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return errors.Wrap(err, "unable to create configmap")
			}
		}
	}
	return nil
}

// CreateOrUpdateSecret creates a Secret if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateSecret(client clientset.Interface, secret *corev1.Secret) error {
	if _, err := client.CoreV1().Secrets(secret.ObjectMeta.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create secret")
		}

		if _, err := client.CoreV1().Secrets(secret.ObjectMeta.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update secret")
		}
	}
	return nil
}

// CreateOrUpdateServiceAccount creates a ServiceAccount if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateServiceAccount(client clientset.Interface, sa *corev1.ServiceAccount) error {
	if _, err := client.CoreV1().ServiceAccounts(sa.ObjectMeta.Namespace).Create(context.TODO(), sa, metav1.CreateOptions{}); err != nil {
		// Note: We don't run .Update here afterwards as that's probably not required
		// Only thing that could be updated is annotations/labels in .metadata, but we don't use that currently
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create serviceaccount")
		}
	}
	return nil
}

// CreateOrUpdateDeployment creates a Deployment if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateDeployment(client clientset.Interface, deploy *apps.Deployment) error {
	if _, err := client.AppsV1().Deployments(deploy.ObjectMeta.Namespace).Create(context.TODO(), deploy, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create deployment")
		}

		if _, err := client.AppsV1().Deployments(deploy.ObjectMeta.Namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update deployment")
		}
	}
	return nil
}

// CreateOrUpdateDaemonSet creates a DaemonSet if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateDaemonSet(client clientset.Interface, ds *apps.DaemonSet) error {
	if _, err := client.AppsV1().DaemonSets(ds.ObjectMeta.Namespace).Create(context.TODO(), ds, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create daemonset")
		}

		if _, err := client.AppsV1().DaemonSets(ds.ObjectMeta.Namespace).Update(context.TODO(), ds, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update daemonset")
		}
	}
	return nil
}

// DeleteDaemonSetForeground deletes the specified DaemonSet in foreground mode; i.e. it blocks until/makes sure all the managed Pods are deleted
func DeleteDaemonSetForeground(client clientset.Interface, namespace, name string) error {
	foregroundDelete := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &foregroundDelete,
	}
	return client.AppsV1().DaemonSets(namespace).Delete(context.TODO(), name, deleteOptions)
}

// DeleteDeploymentForeground deletes the specified Deployment in foreground mode; i.e. it blocks until/makes sure all the managed Pods are deleted
func DeleteDeploymentForeground(client clientset.Interface, namespace, name string) error {
	foregroundDelete := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &foregroundDelete,
	}
	return client.AppsV1().Deployments(namespace).Delete(context.TODO(), name, deleteOptions)
}

// CreateOrUpdateRole creates a Role if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateRole(client clientset.Interface, role *rbac.Role) error {
	if _, err := client.RbacV1().Roles(role.ObjectMeta.Namespace).Create(context.TODO(), role, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create RBAC role")
		}

		if _, err := client.RbacV1().Roles(role.ObjectMeta.Namespace).Update(context.TODO(), role, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update RBAC role")
		}
	}
	return nil
}

// CreateOrUpdateRoleBinding creates a RoleBinding if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateRoleBinding(client clientset.Interface, roleBinding *rbac.RoleBinding) error {
	if _, err := client.RbacV1().RoleBindings(roleBinding.ObjectMeta.Namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create RBAC rolebinding")
		}

		if _, err := client.RbacV1().RoleBindings(roleBinding.ObjectMeta.Namespace).Update(context.TODO(), roleBinding, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update RBAC rolebinding")
		}
	}
	return nil
}

// CreateOrUpdateClusterRole creates a ClusterRole if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateClusterRole(client clientset.Interface, clusterRole *rbac.ClusterRole) error {
	if _, err := client.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create RBAC clusterrole")
		}

		if _, err := client.RbacV1().ClusterRoles().Update(context.TODO(), clusterRole, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update RBAC clusterrole")
		}
	}
	return nil
}

// CreateOrUpdateClusterRoleBinding creates a ClusterRoleBinding if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateClusterRoleBinding(client clientset.Interface, clusterRoleBinding *rbac.ClusterRoleBinding) error {
	if _, err := client.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create RBAC clusterrolebinding")
		}

		if _, err := client.RbacV1().ClusterRoleBindings().Update(context.TODO(), clusterRoleBinding, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update RBAC clusterrolebinding")
		}
	}
	return nil
}

// PatchNodeOnce executes patchFn on the node object found by the node name.
// This is a condition function meant to be used with wait.Poll. false, nil
// implies it is safe to try again, an error indicates no more tries should be
// made and true indicates success.
func PatchNodeOnce(client clientset.Interface, nodeName string, patchFn func(*corev1.Node)) func() (bool, error) {
	return func() (bool, error) {
		// First get the node object
		n, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		// The node may appear to have no labels at first,
		// so we wait for it to get hostname label.
		if _, found := n.ObjectMeta.Labels[LabelHostname]; !found {
			return false, nil
		}

		oldData, err := json.Marshal(n)
		if err != nil {
			return false, errors.Wrapf(err, "failed to marshal unmodified node %q into JSON", n.Name)
		}

		// Execute the mutating function
		patchFn(n)

		newData, err := json.Marshal(n)
		if err != nil {
			return false, errors.Wrapf(err, "failed to marshal modified node %q into JSON", n.Name)
		}

		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Node{})
		if err != nil {
			return false, errors.Wrap(err, "failed to create two way merge patch")
		}

		if _, err := client.CoreV1().Nodes().Patch(context.TODO(), n.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
			if apierrors.IsConflict(err) {
				fmt.Println("[patchnode] Temporarily unable to update node metadata due to conflict (will retry)")
				return false, nil
			}
			return false, errors.Wrapf(err, "error patching node %q through apiserver", n.Name)
		}

		return true, nil
	}
}

// PatchNode tries to patch a node using patchFn for the actual mutating logic.
// Retries are provided by the wait package.
func PatchNode(client clientset.Interface, nodeName string, patchFn func(*corev1.Node)) error {
	// wait.Poll will rerun the condition function every interval function if
	// the function returns false. If the condition function returns an error
	// then the retries end and the error is returned.
	return wait.Poll(APICallRetryInterval, PatchNodeTimeout, PatchNodeOnce(client, nodeName, patchFn))
}

// CreateOrUpdateService creates a service if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateService(client clientset.Interface, svc *corev1.Service) error {
	svcLive, err := client.CoreV1().Services(svc.ObjectMeta.Namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
	if err == nil {
		svc.ResourceVersion = svcLive.ResourceVersion
		_, err := client.CoreV1().Services(svc.ObjectMeta.Namespace).Update(context.TODO(), svc, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	} else {
		if _, err := client.CoreV1().Services(svc.ObjectMeta.Namespace).Create(context.TODO(), svc, metav1.CreateOptions{}); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return errors.Wrap(err, "unable to create service")
			}

			if _, err := client.CoreV1().Services(svc.ObjectMeta.Namespace).Update(context.TODO(), svc, metav1.UpdateOptions{}); err != nil {
				return errors.Wrap(err, "unable to update service")
			}
		}
	}

	return nil
}

// CreateOrUpdateStatefulSet creates a statefulSet if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateStatefulSet(client clientset.Interface, sts *apps.StatefulSet) error {
	if _, err := client.AppsV1().StatefulSets(sts.ObjectMeta.Namespace).Create(context.TODO(), sts, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create statefulSet")
		}

		if _, err := client.AppsV1().StatefulSets(sts.ObjectMeta.Namespace).Update(context.TODO(), sts, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update statefulSet")
		}
	}
	return nil
}

// CreateOrUpdateNamespace creates a namespace if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateNamespace(client clientset.Interface, ns *corev1.Namespace) error {
	if _, err := client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create namespace")
		}

		if _, err := client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update namespace")
		}
	}
	return nil
}

// CreateOrUpdateEndpoints creates a Endpoints if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateEndpoints(client clientset.Interface, ep *corev1.Endpoints) error {
	if _, err := client.CoreV1().Endpoints(ep.ObjectMeta.Namespace).Create(context.TODO(), ep, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create endpoints")
		}

		if _, err := client.CoreV1().Endpoints(ep.ObjectMeta.Namespace).Update(context.TODO(), ep, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update endpoints")
		}
	}
	return nil
}

// CreateOrUpdateIngress creates a Ingress if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateIngress(client clientset.Interface, ing *extensionsv1beta1.Ingress) error {
	if _, err := client.ExtensionsV1beta1().Ingresses(ing.ObjectMeta.Namespace).Create(context.TODO(), ing, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create ingress")
		}

		if _, err := client.ExtensionsV1beta1().Ingresses(ing.ObjectMeta.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update ingress")
		}
	}
	return nil
}

// CreateOrUpdateIngress creates a podSecurityPolicy if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdatePodSecurityPolicy(client clientset.Interface, podSecurityPolicy *v1beta1.PodSecurityPolicy) error {
	client.PolicyV1beta1().PodSecurityPolicies().Delete(context.TODO(), podSecurityPolicy.Name, metav1.DeleteOptions{})
	if _, err := client.PolicyV1beta1().PodSecurityPolicies().Create(context.TODO(), podSecurityPolicy, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create podSecurityPolicy")
		}
		if _, err := client.PolicyV1beta1().PodSecurityPolicies().Update(context.TODO(), podSecurityPolicy, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update podSecurityPolicy")
		}
	}
	return nil
}

func CreateOrUpdateCSIDriver(client clientset.Interface, csiDriver *storagev1.CSIDriver) error {
	if _, err := client.StorageV1().CSIDrivers().Create(context.TODO(), csiDriver, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create podSecurityPolicy")
		}
		if _, err := client.StorageV1().CSIDrivers().Update(context.TODO(), csiDriver, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update podSecurityPolicy")
		}
	}
	return nil
}

func CreateOrUpdateStorageClass(client clientset.Interface, csiDriver *storagev1.StorageClass) error {
	if _, err := client.StorageV1().StorageClasses().Create(context.TODO(), csiDriver, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create podSecurityPolicy")
		}
		if _, err := client.StorageV1().StorageClasses().Update(context.TODO(), csiDriver, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update podSecurityPolicy")
		}
	}
	return nil
}

// CreateOrUpdateJob creates a Job if the target resource doesn't exist. If the resource exists already, this function will update
func CreateOrUpdateJob(client clientset.Interface, job *batchv1.Job) error {
	if _, err := client.BatchV1().Jobs(job.ObjectMeta.Namespace).Create(context.TODO(), job, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create job")
		}

		if _, err := client.BatchV1().Jobs(job.ObjectMeta.Namespace).Update(context.TODO(), job, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update job")
		}
	}
	return nil
}

// CreateOrUpdateCronJob creates a Job if the target resource doesn't exist. If the resource exists already, this function will update
func CreateOrUpdateCronJob(client clientset.Interface, cronjob *batchv1beta1.CronJob) error {
	if _, err := client.BatchV1beta1().CronJobs(cronjob.ObjectMeta.Namespace).Create(context.TODO(), cronjob, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create cronjob")
		}

		if _, err := client.BatchV1beta1().CronJobs(cronjob.ObjectMeta.Namespace).Update(context.TODO(), cronjob, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update cronjob")
		}
	}
	return nil
}

// CreateOrUpdateConfigMapFromFile like kubectl create configmap --from-file
func CreateOrUpdateConfigMapFromFile(client clientset.Interface, cm *corev1.ConfigMap, pattern string) error {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return errors.New("no matches found")
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	for _, filename := range matches {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		cm.Data[filepath.Base(filename)] = string(data)
	}

	return CreateOrUpdateConfigMap(client, cm)
}

func int32Ptr(i int32) *int32 {
	o := i
	return &o
}

// DeleteReplicaSetApp delete the replicaset and pod additionally for deployment app with extension group
func DeleteReplicaSetApp(client clientset.Interface, options metav1.ListOptions) error {
	rsList, err := client.ExtensionsV1beta1().ReplicaSets(metav1.NamespaceSystem).List(context.TODO(), options)
	if err != nil {
		return err
	}

	var errs []error
	for i := range rsList.Items {
		rs := &rsList.Items[i]
		// update replicas to zero
		rs.Spec.Replicas = int32Ptr(0)
		_, err = client.ExtensionsV1beta1().ReplicaSets(metav1.NamespaceSystem).Update(context.TODO(), rs, metav1.UpdateOptions{})
		if err != nil {
			errs = append(errs, err)
		} else {
			// delete replicaset
			err = client.ExtensionsV1beta1().ReplicaSets(metav1.NamespaceSystem).Delete(context.TODO(), rs.Name, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		errMsg := ""
		for _, e := range errs {
			errMsg += e.Error() + ";"
		}
		return fmt.Errorf("delete replicaSet fail:%s", errMsg)
	}

	return nil
}

// CreateOrUpdateDeploymentExtensionsV1beta1 creates a ExtensionsV1beta1 Deployment if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateDeploymentExtensionsV1beta1(client clientset.Interface, deploy *extensionsv1beta1.Deployment) error {
	if _, err := client.ExtensionsV1beta1().Deployments(deploy.ObjectMeta.Namespace).Create(context.TODO(), deploy, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create deployment")
		}

		if _, err := client.ExtensionsV1beta1().Deployments(deploy.ObjectMeta.Namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update deployment")
		}
	}
	return nil
}

// DeleteExtensionsV1beta1Deployment delete a deployment
func deleteDeploymentExtensionsV1beta1(client clientset.Interface, deployName string, labelSelector metav1.LabelSelector) error {
	retErr := client.ExtensionsV1beta1().Deployments(metav1.NamespaceSystem).Delete(context.TODO(), deployName, metav1.DeleteOptions{})

	// Delete replicaset for extensions groups
	if retErr != nil {
		selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
		if err != nil {
			retErr = err
		} else {
			options := metav1.ListOptions{
				LabelSelector: selector.String(),
			}
			err = DeleteReplicaSetApp(client, options)
			if err != nil {
				retErr = err
			}
		}
	}
	return retErr
}

// DeleteDeployment delete a deployment
func deleteDeploymentAppsV1(client clientset.Interface, namespace string, deployName string) error {
	return client.AppsV1().Deployments(namespace).Delete(context.TODO(), deployName, metav1.DeleteOptions{})
}

// DeleteServiceAccounts delete a serviceAccount
func DeleteServiceAccounts(client clientset.Interface, namespace string, svcAccountName string) error {
	return client.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), svcAccountName, metav1.DeleteOptions{})
}

// GetService get a service.
func GetService(client clientset.Interface, namespace string, name string) (*corev1.Service, error) {
	return client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetServiceAccount get a service.
func GetServiceAccount(client clientset.Interface, namespace string, name string) (*corev1.ServiceAccount, error) {
	return client.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetClusterRoleBinding get a cluster role binding.
func GetClusterRoleBinding(client clientset.Interface, name string) (*rbac.ClusterRoleBinding, error) {
	return client.RbacV1().ClusterRoleBindings().Get(context.TODO(), name, metav1.GetOptions{})
}

// CreateOrUpdateValidatingWebhookConfiguration creates a ValidatingWebhookConfigurations if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateValidatingWebhookConfiguration(client clientset.Interface, obj *admissionv1.ValidatingWebhookConfiguration) error {
	if _, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.TODO(), obj, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create ValidatingWebhookConfiguration")
		}

		if _, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update ValidatingWebhookConfiguration")
		}
	}

	return nil
}

// CreateOrUpdateMutatingWebhookConfiguration creates a MutatingWebhookConfigurations if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdateMutatingWebhookConfiguration(client clientset.Interface, obj *admissionv1.MutatingWebhookConfiguration) error {
	client.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.TODO(), obj.Name, metav1.DeleteOptions{})
	if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), obj, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create MutatingWebhookConfiguration")
		}

		if _, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update MutatingWebhookConfiguration")
		}
	}

	return nil
}

// CreateOrUpdatePersistentVolume creates a PersistentVolume if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func CreateOrUpdatePersistentVolume(client clientset.Interface, obj *corev1.PersistentVolume) error {
	if _, err := client.CoreV1().PersistentVolumes().Get(context.TODO(), obj.Name, metav1.GetOptions{}); err == nil {
		// pv should not update
		return nil
	}
	if _, err := client.CoreV1().PersistentVolumes().Create(context.TODO(), obj, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create PersistentVolume")
		}

		if _, err := client.CoreV1().PersistentVolumes().Update(context.TODO(), obj, metav1.UpdateOptions{}); err != nil {
			return errors.Wrap(err, "unable to update PersistentVolume")
		}
	}

	return nil
}

// MarkNode mark node by adding labels and taints
func MarkNode(client clientset.Interface, nodeName string, labels map[string]string, taints []corev1.Taint) error {
	return PatchNode(client, nodeName, func(n *corev1.Node) {
		for k, v := range labels {
			n.Labels[k] = v
		}

		for _, oldTaint := range n.Spec.Taints {
			existed := false
			for _, newTaint := range taints {
				if newTaint.MatchTaint(&oldTaint) {
					existed = true
					break
				}
			}
			if !existed {
				taints = append(taints, oldTaint)
			}
		}
		n.Spec.Taints = taints
	})
}

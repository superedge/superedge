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
	"reflect"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

var deleteHandlers map[string]func(kubernetes.Interface, []byte) error

func init() {
	deleteHandlers = make(map[string]func(kubernetes.Interface, []byte) error)
	deleteHandlers["Job"] = deleteJob
	deleteHandlers["Role"] = deleteRole
	deleteHandlers["Secret"] = deleteSecret
	deleteHandlers["Service"] = deleteService
	deleteHandlers["DaemonSet"] = deleteDaemonSet
	deleteHandlers["ConfigMap"] = deleteConfigMap
	deleteHandlers["CSIDriver"] = deleteCSIDriver
	deleteHandlers["Deployment"] = deleteDeployment
	deleteHandlers["RoleBinding"] = deleteRoleBinding
	deleteHandlers["ClusterRole"] = deleteClusterRole
	deleteHandlers["StorageClass"] = deleteStorageClass
	deleteHandlers["ServiceAccount"] = deleteServiceAccount
	deleteHandlers["PodSecurityPolicy"] = deletePodSecurityPolicy
	deleteHandlers["ClusterRoleBinding"] = deleteClusterRoleBinding
	deleteHandlers["MutatingWebhookConfiguration"] = deleteMutatingMutatingWebhookConfigurations
}

func deleteConfigMap(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.ConfigMap)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteConfigMap(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteSecret(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.Secret)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteSecret(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteService(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.Service)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteService(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteServiceAccount(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.ServiceAccount)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteServiceAccount(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteDaemonSet(client kubernetes.Interface, data []byte) error {
	obj := new(appsv1.DaemonSet)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteDaemonSet(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteDeployment(client kubernetes.Interface, data []byte) error {
	obj := new(appsv1.Deployment)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteDeployment(client, obj.Namespace, obj.Name, false, metav1.LabelSelector{})
	if err != nil {
		return err
	}
	return nil
}

func deleteRole(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.Role)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteRole(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteRoleBinding(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.RoleBinding)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteRoleBinding(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteClusterRole(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.ClusterRole)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteClusterRole(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteClusterRoleBinding(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.ClusterRoleBinding)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteClusterRoleBinding(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteCSIDriver(client kubernetes.Interface, data []byte) error {
	obj := new(storagev1.CSIDriver)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteCSIDriver(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteStorageClass(client kubernetes.Interface, data []byte) error {
	obj := new(storagev1.StorageClass)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteStorageClasses(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deletePodSecurityPolicy(client kubernetes.Interface, data []byte) error {
	obj := new(v1beta1.PodSecurityPolicy)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeletePodSecurityPolicy(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteMutatingMutatingWebhookConfigurations(client kubernetes.Interface, data []byte) error {
	obj := new(admissionv1.MutatingWebhookConfiguration)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteMutatingMutatingWebhookConfigurations(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func deleteJob(client kubernetes.Interface, data []byte) error {
	obj := new(batchv1.Job)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := DeleteJob(client, obj)
	if err != nil {
		return err
	}
	return nil
}

// DeleteResourceWithFile create k8s resource with file
func DeleteResourceWithFile(client kubernetes.Interface, yamlStr string, option interface{}) error {
	var (
		data []byte
		err  error
	)

	if option == nil {
		option = map[string]interface{}{}
	}

	data, err = ParseString(yamlStr, option)
	if err != nil {
		return err
	}

	klog.V(6).Infof("Delete Kubernetes resource yaml: %s", string(data))

	reg := regexp.MustCompile(`(?m)^-{3,}$`)
	items := reg.Split(string(data), -1)
	for _, item := range items {
		objBytes := []byte(item)
		obj := new(object)
		err := yaml.Unmarshal(objBytes, obj)
		if err != nil {
			return err
		}
		if obj.Kind == "" {
			continue
		}
		f, ok := deleteHandlers[obj.Kind]
		if !ok {
			return errors.Errorf("unsupport kind %q", obj.Kind)
		}
		err = f(client, objBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

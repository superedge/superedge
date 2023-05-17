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
	"bytes"
	"reflect"
	"regexp"
	"text/template"

	"k8s.io/klog/v2"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

var handlers map[string]func(kubernetes.Interface, []byte) error

func init() {
	handlers = make(map[string]func(kubernetes.Interface, []byte) error)
	handlers["Job"] = createOrUpdateJob
	handlers["Role"] = createOrUpdateRole
	handlers["Secret"] = createOrUpdateSecret
	handlers["Service"] = createOrUpdateService
	handlers["CronJob"] = createOrUpdateCronJob
	handlers["Ingress"] = createOrUpdateIngress
	handlers["CSIDriver"] = createOrUpdateCSIDriver
	handlers["Endpoints"] = createOrUpdateEndpoints
	handlers["Namespace"] = createOrUpdateNamespace
	handlers["DaemonSet"] = createOrUpdateDaemonSet
	handlers["ConfigMap"] = createOrUpdateConfigMap
	handlers["Deployment"] = createOrUpdateDeployment
	handlers["StatefulSet"] = createOrUpdateStatefulSet
	handlers["RoleBinding"] = createOrUpdateRoleBinding
	handlers["ClusterRole"] = createOrUpdateClusterRole
	handlers["StorageClass"] = createOrUpdateStorageClass
	handlers["ServiceAccount"] = createOrUpdateServiceAccount
	handlers["PodSecurityPolicy"] = createOrUpdatePodSecurityPolicy
	handlers["ClusterRoleBinding"] = createOrUpdateClusterRoleBinding
	handlers["MutatingWebhookConfiguration"] = createOrUpdateMutatingWebhookConfiguration
	handlers["ValidatingWebhookConfiguration"] = createOrUpdateValidatingWebhookConfiguration
	handlers["PersistentVolume"] = createOrUpdatePersistentVolume
}

func createOrUpdateConfigMap(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.ConfigMap)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateConfigMap(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateEndpoints(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.Endpoints)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateEndpoints(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateDeployment(client kubernetes.Interface, data []byte) error {
	obj := new(appsv1.Deployment)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateDeployment(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateNamespace(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.Namespace)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateNamespace(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateSecret(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.Secret)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateSecret(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateService(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.Service)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateService(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateServiceAccount(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.ServiceAccount)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateServiceAccount(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateJob(client kubernetes.Interface, data []byte) error {
	obj := new(batchv1.Job)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateJob(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateCronJob(client kubernetes.Interface, data []byte) error {
	obj := new(batchv1beta1.CronJob)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateCronJob(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateDaemonSet(client kubernetes.Interface, data []byte) error {
	obj := new(appsv1.DaemonSet)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateDaemonSet(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateStatefulSet(client kubernetes.Interface, data []byte) error {
	obj := new(appsv1.StatefulSet)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateStatefulSet(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateIngress(client kubernetes.Interface, data []byte) error {
	obj := new(extensionsv1beta1.Ingress)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateIngress(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdatePodSecurityPolicy(client kubernetes.Interface, data []byte) error {
	obj := new(v1beta1.PodSecurityPolicy)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdatePodSecurityPolicy(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateCSIDriver(client kubernetes.Interface, data []byte) error {
	obj := new(storagev1.CSIDriver)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateCSIDriver(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateStorageClass(client kubernetes.Interface, data []byte) error {
	obj := new(storagev1.StorageClass)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateStorageClass(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateRole(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.Role)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateRole(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateRoleBinding(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.RoleBinding)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateRoleBinding(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateClusterRole(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.ClusterRole)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateClusterRole(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateClusterRoleBinding(client kubernetes.Interface, data []byte) error {
	obj := new(rbacv1.ClusterRoleBinding)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateClusterRoleBinding(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateValidatingWebhookConfiguration(client kubernetes.Interface, data []byte) error {
	obj := new(admissionv1.ValidatingWebhookConfiguration)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateValidatingWebhookConfiguration(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdateMutatingWebhookConfiguration(client kubernetes.Interface, data []byte) error {
	obj := new(admissionv1.MutatingWebhookConfiguration)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdateMutatingWebhookConfiguration(client, obj)
	if err != nil {
		return err
	}
	return nil
}

func createOrUpdatePersistentVolume(client kubernetes.Interface, data []byte) error {
	obj := new(corev1.PersistentVolume)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err := CreateOrUpdatePersistentVolume(client, obj)
	if err != nil {
		return err
	}
	return nil
}

// ParseString validates and parses passed as argument template
func ParseString(strtmpl string, obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	tmpl, err := template.New("template").Parse(strtmpl)
	if err != nil {
		return nil, errors.Wrap(err, "error when parsing template")
	}
	err = tmpl.Execute(&buf, obj)
	if err != nil {
		return nil, errors.Wrap(err, "error when executing template")
	}
	return buf.Bytes(), nil
}

// CompleteTemplate complete templates with context
func CompleteTemplate(strtmpl string, context interface{}) (string, error) {
	writer := bytes.NewBuffer([]byte{})
	tmpl, err := template.New("template").Option("missingkey=zero").Parse(strtmpl)
	if err != nil {
		return "", errors.Wrap(err, "error when parsing template")
	}
	err = tmpl.Execute(writer, context)
	if nil != err {
		return "", errors.Wrap(err, "error when executing template")
	}
	return writer.String(), nil
}

type object struct {
	Kind string `yaml:"kind"`
}

// CreateResourceWithFile create k8s resource with file
func CreateResourceWithFile(client kubernetes.Interface, yamlStr string, option interface{}) error {
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

	klog.V(6).Infof("Create kubernetes resource output yaml: %s", string(data))

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
		f, ok := handlers[obj.Kind]
		if !ok {
			return errors.Errorf("unsupport kind %q", obj.Kind)
		}
		err = f(client, objBytes)
		if err != nil {
			klog.Errorf("Apply %s error: %v", obj.Kind, err)
			return err
		}
	}

	return nil
}

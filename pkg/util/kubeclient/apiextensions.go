package kubeclient

import (
	"context"
	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionSclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"reflect"
)

func GetAPIExtensionSclientset(kubeconfigFile string) (extensionSclientset.Interface, error) {
	cfg, err := GetKubeConfig(kubeconfigFile)
	if err != nil {
		klog.Errorf("GetClientSet error: %v", err)
		return nil, err
	}

	client, err := extensionSclientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to build apiextensions clientset, error:%v", err)
		return nil, err
	}

	return client, nil
}

func CreateOrUpdateCustomResourceDefinition(client extensionSclientset.Interface, yaml string, option interface{}) error {
	data, err := ParseString(yaml, option)
	if err != nil {
		return err
	}

	obj := new(apiextensionsv1.CustomResourceDefinition)
	if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), data, obj); err != nil {
		return errors.Wrapf(err, "unable to decode %s", reflect.TypeOf(obj).String())
	}
	err = createOrUpdateCustomResourceDefinition(client, obj)
	if err != nil {
		return err
	}
	return nil
}

// CreateOrUpdateCustomResourceDefinition creates a CustomResourceDefinition if the target resource doesn't exist. If the resource exists already, this function will update the resource instead.
func createOrUpdateCustomResourceDefinition(client extensionSclientset.Interface, crd *apiextensionsv1.CustomResourceDefinition) error {
	if _, err := client.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create crd")
		}

		crdOld, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crd.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to update crd")
		}

		crd.ResourceVersion = crdOld.ResourceVersion
		_, err = client.ApiextensionsV1().CustomResourceDefinitions().Update(context.TODO(), crd, metav1.UpdateOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to update crd")
		}
	}
	return nil
}

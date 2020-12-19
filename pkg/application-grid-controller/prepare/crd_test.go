package prepare

import (
	"superedge/pkg/application-grid-controller/controller/common"
	"superedge/pkg/util/kubeclient"
	"gopkg.in/yaml.v2"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	kuberuntime "k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"regexp"
	"testing"
)

func TestYaml(t *testing.T) {
	data, err := kubeclient.ParseString(common.DeploymentGridCRDYaml, map[string]interface{}{})
	if err != nil {
		t.Error("err")
	}
	//klog.V(8).Infof("Create yaml: %s", string(data))
	reg := regexp.MustCompile(`(?m)^-{3,}$`)
	items := reg.Split(string(data), -1)
	for _, item := range items {
		objBytes := []byte(item)
		obj := new(object)
		err := yaml.Unmarshal(objBytes, obj)
		if err != nil {
			t.Error("err")
		}
		if obj.Kind == "" {
			continue
		}
		objcrd := new(apiext.CustomResourceDefinition)
		if err := kuberuntime.DecodeInto(clientsetscheme.Codecs.UniversalDecoder(), objBytes, objcrd); err != nil {
			t.Error("err")
		}
		t.Logf("%v", objcrd.Spec.Validation.OpenAPIV3Schema.Properties)
	}
}

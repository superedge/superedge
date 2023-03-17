package webhook

import (
	"github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io"
	sitev1alpha1 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1"
	sitev1alpha2 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var scheme = runtime.NewScheme()
var codec = serializer.NewCodecFactory(scheme)

var apiextensionsLocalSchemeBuilder = runtime.SchemeBuilder{
	apiextensionsv1beta1.AddToScheme,
	apiextensionsv1.AddToScheme,
}

var siteLocalSchemeBuilder = runtime.SchemeBuilder{
	sitev1alpha1.AddToScheme,
	sitev1alpha2.AddToScheme,
	site.AddToScheme,
}

var apiextensionsAddToScheme = apiextensionsLocalSchemeBuilder.AddToScheme
var siteAddToScheme = siteLocalSchemeBuilder.AddToScheme

func init() {
	utilruntime.Must(apiextensionsAddToScheme(scheme))
	utilruntime.Must(siteAddToScheme(scheme))
}

func init() {
	v1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})

}

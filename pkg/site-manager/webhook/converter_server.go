package webhook

import (
	"encoding/json"
	"io/ioutil"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

func V1Handler(writer http.ResponseWriter, request *http.Request) {

	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		klog.Error(err)
		return
	}

	convReview := &apiextensionsv1.ConversionReview{}
	convResp := &apiextensionsv1.ConversionResponse{
		ConvertedObjects: []runtime.RawExtension{},
	}

	deserializer := codec.UniversalDeserializer()
	_, _, err = deserializer.Decode(data, nil, convReview)
	if err != nil {
		klog.Error(err)
		return
	}
	convResp.UID = convReview.Request.UID
	convReview.Response = convResp
	for _, v := range convReview.Request.Objects {
		obj, err := convert(v.Raw, convReview.Request.DesiredAPIVersion)
		if err != nil {
			klog.Error(err)
			convResp.Result = metav1.Status{
				Status:  metav1.StatusFailure,
				Message: err.Error(),
				Reason:  metav1.StatusReasonInternalError,
				Code:    http.StatusInternalServerError,
			}

			writeObject(convReview, writer)
			return
		}
		convResp.ConvertedObjects = append(convResp.ConvertedObjects, runtime.RawExtension{Object: obj})
	}
	convResp.Result = metav1.Status{
		Status: metav1.StatusSuccess,
	}
	writeObject(convReview, writer)
}

func convert(data []byte, desiredAPIVersion string) (runtime.Object, error) {
	srcObj := &unstructured.Unstructured{}
	_, gvk, err := codec.UniversalDeserializer().Decode(data, nil, srcObj)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if gvk.GroupVersion().String() == desiredAPIVersion {
		return srcObj, nil
	}

	gvk.Version = runtime.APIVersionInternal
	internal, err := scheme.New(*gvk)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	err = scheme.Convert(srcObj, internal, gvk.GroupVersion())
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	gvk.Version = strings.Split(desiredAPIVersion, "/")[1]
	target, err := scheme.New(*gvk)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	err = scheme.Convert(internal, target, gvk.Version)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	scheme.Default(target)
	return target, nil
}

func writeObject(object *apiextensionsv1.ConversionReview, writer http.ResponseWriter) {

	r, err := json.Marshal(object)
	if err != nil {
		_, err = writer.Write([]byte("Failed to marshal"))
		if err != nil {
			klog.Error(err)
			return
		}
		return
	}

	_, err = writer.Write(r)
	if err != nil {
		klog.Error(err)
		return
	}
}

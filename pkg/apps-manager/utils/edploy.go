package utils

import (
	"context"
	"fmt"
	appsv1 "github.com/superedge/superedge/pkg/apps-manager/apis/apps/v1"
	"github.com/superedge/superedge/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"strconv"
)

func SchedulableNode(kubeClient clientset.Interface, edeploy *appsv1.EDeployment) ([]corev1.Node, error) {
	podTemplate := edeploy.Spec.Template
	if podTemplate.Spec.NodeSelector == nil {
		return nil, fmt.Errorf("Edeploy: %s not fit node\n", edeploy.Name)
	}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: podTemplate.Spec.NodeSelector,
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, err
	}

	listOptions := metav1.ListOptions{LabelSelector: selector.String()}
	nodeList, err := kubeClient.CoreV1().Nodes().List(context.TODO(), listOptions)
	if err != nil {
		klog.Errorf("Get nodes by selector: %s, error: %v", selector.String(), err)
		return nil, err
	}

	return nodeList.Items, nil
}

func WriteEdeployToStaticPod(kubeClient clientset.Interface, edeploy *appsv1.EDeployment, staticPodDir string) error {
	replicas := *edeploy.Spec.Replicas
	podTemplate := &edeploy.Spec.Template
	for i := 1; i < int(replicas)+1; i++ {
		podTemplate.Name = podTemplate.Name + "-" + strconv.Itoa(int(i))
		staticPod := &corev1.Pod{}
		staticPod.TypeMeta = metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		}
		podTemplate.Spec.DeepCopyInto(&staticPod.Spec)
		podTemplate.ObjectMeta.DeepCopyInto(&staticPod.ObjectMeta)
		podData, err := util.PodToYaml(staticPod)
		if err = util.WriteWithBufio(staticPodDir, string(podData)); err != nil {
			klog.Errorf("Write file: %s error: %v", staticPodDir, err)
			continue
		}
	}
	return nil
}

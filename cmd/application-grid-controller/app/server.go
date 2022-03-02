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

package app

import (
	"context"
	"k8s.io/component-base/featuregate"
	"os"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
	"github.com/superedge/superedge/cmd/application-grid-controller/app/options"
	superedge "github.com/superedge/superedge/pkg/application-grid-controller/apis/superedge.io"
	"github.com/superedge/superedge/pkg/application-grid-controller/config"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/deployment"
	deploymentutil "github.com/superedge/superedge/pkg/application-grid-controller/controller/deployment/util"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/service"
	serviceutil "github.com/superedge/superedge/pkg/application-grid-controller/controller/service/util"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/statefulset"
	statefulsetutil "github.com/superedge/superedge/pkg/application-grid-controller/controller/statefulset/util"
	crdClientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned/scheme"
	"github.com/superedge/superedge/pkg/application-grid-controller/prepare"
	controllerutil "github.com/superedge/superedge/pkg/application-grid-controller/util"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
)

func NewGridControllerManagerCommand() *cobra.Command {
	o := options.NewApplicationGridControllerOptions()

	cmd := &cobra.Command{
		Use: "application-grid-controller",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			//if err := utilfeature.DefaultMutableFeatureGate.SetFromMap(o.FeatureGates); err != nil {
			//	klog.Errorf("failed to set feature gate, %v", err)
			//}
			var parentCrdClient *crdClientset.Clientset
			var once sync.Once
			var electionChecker *leaderelection.HealthzAdaptor

			kubeconfig, err := clientcmd.BuildConfigFromFlags(o.Master, o.Kubeconfig)
			if err != nil {
				klog.Fatalf("failed to create kubeconfig: %v", err)
			}

			kubeconfig.QPS = o.QPS
			kubeconfig.Burst = o.Burst

			copyConfig := *kubeconfig
			copyConfig.Timeout = time.Second * o.RenewDeadline.Duration

			leaderElectionClient := clientset.NewForConfigOrDie(restclient.AddUserAgent(&copyConfig, "leader-election"))
			kubeClient := clientset.NewForConfigOrDie(kubeconfig)
			crdClient := crdClientset.NewForConfigOrDie(kubeconfig)
			apiextensionClient := apiextclientset.NewForConfigOrDie(kubeconfig)

			parentKubeconfig, dedicatedNameSpace, err := generateParentDedicatedKubeConfig(kubeClient)
			if err != nil {
				klog.Fatalf("failed to create kubeconfig: %v", err)
			}
			if parentKubeconfig != nil {
				parentCrdClient = crdClientset.NewForConfigOrDie(parentKubeconfig)
			}

			once.Do(func() {
				if o.FeatureGates["EndpointSlice"] == true {
					common.DefaultKubernetesFeatureGates[common.EndpointSlice] = featuregate.FeatureSpec{Default: true, PreRelease: featuregate.Beta}
				}
				runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(common.DefaultKubernetesFeatureGates))
			})

			if !o.LeaderElect {
				runController(context.TODO(), apiextensionClient, kubeClient, crdClient, parentCrdClient, o.Worker, o.SyncPeriod, dedicatedNameSpace)
				panic("unreachable")
			}

			electionChecker = leaderelection.NewLeaderHealthzAdaptor(time.Second * 20)

			id, err := os.Hostname()
			if err != nil {
				klog.Fatalf("failed to get hostname %v", err)
			}
			id = id + "_" + string(uuid.NewUUID())
			eventRecorder := createRecorder(kubeClient, options.ApplicationGridControllerUserAgent)
			rl, err := resourcelock.New(o.ResourceLock, o.ResourceNamespace, o.ResourceName,
				leaderElectionClient.CoreV1(),
				leaderElectionClient.CoordinationV1(),
				resourcelock.ResourceLockConfig{
					Identity:      id,
					EventRecorder: eventRecorder,
				})
			if err != nil {
				klog.Fatalf("error creating lock: %v", err)
			}

			leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
				Lock:          rl,
				LeaseDuration: o.LeaseDuration.Duration,
				RenewDeadline: o.RenewDeadline.Duration,
				RetryPeriod:   o.RetryPeriod.Duration,
				Callbacks: leaderelection.LeaderCallbacks{
					OnStartedLeading: func(ctx context.Context) {
						runController(ctx, apiextensionClient, kubeClient, crdClient, parentCrdClient, o.Worker, o.SyncPeriod, dedicatedNameSpace)
					},
					OnStoppedLeading: func() {
						klog.Fatalf("leaderelection lost")
					},
				},
				WatchDog: electionChecker,
				Name:     options.ApplicationGridControllerUserAgent,
			})
			panic("unreachable")
		},
	}

	fs := cmd.Flags()
	o.AddFlags(fs)

	return cmd
}

func runController(parent context.Context,
	apiextensionClient *apiextclientset.Clientset, kubeClient *clientset.Clientset, crdClient *crdClientset.Clientset, fedCrdClient *crdClientset.Clientset,
	workerNum, syncPeriod int, dedicatedNameSpace string) {

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	// Create and wait for CRDs ready
	crdP := prepare.NewCRDPreparator(apiextensionClient)
	if err := crdP.Prepare(ctx.Done(), schema.GroupVersionKind{
		Group:   superedge.GroupName,
		Version: superedge.Version,
		Kind:    deploymentutil.ControllerKind.Kind,
	}, schema.GroupVersionKind{
		Group:   superedge.GroupName,
		Version: superedge.Version,
		Kind:    statefulsetutil.ControllerKind.Kind,
	}, schema.GroupVersionKind{
		Group:   superedge.GroupName,
		Version: superedge.Version,
		Kind:    serviceutil.ControllerKind.Kind,
	}); err != nil {
		klog.Errorf("Create and wait for CRDs ready failed: %v", err)
	}

	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))

	controllerConfig := config.NewControllerConfig(crdClient, fedCrdClient, kubeClient, time.Second*time.Duration(syncPeriod), dedicatedNameSpace)
	deploymentGridController := deployment.NewDeploymentGridController(
		controllerConfig.DeploymentGridInformer, controllerConfig.DeploymentInformer, controllerConfig.NodeInformer,
		controllerConfig.NameSpaceInformer, kubeClient, crdClient, dedicatedNameSpace)
	statefulSetGridController := statefulset.NewStatefulSetGridController(
		controllerConfig.StatefulSetGridInformer, controllerConfig.StatefulSetInformer, controllerConfig.NodeInformer,
		kubeClient, crdClient)
	serviceGridController := service.NewServiceGridController(controllerConfig.ServiceGridInformer, controllerConfig.ServiceInformer,
		controllerConfig.NameSpaceInformer, kubeClient, crdClient, dedicatedNameSpace)

	var federationdeploymentGridController *deployment.FedDeploymentGridController
	var federationsvcGridController *service.FedServiceGridController
	if fedCrdClient != nil {
		federationdeploymentGridController = deployment.NewFedDeploymentGridController(
			controllerConfig.DeploymentGridInformer, controllerConfig.FedDeploymentGridInformer, controllerConfig.NodeInformer, fedCrdClient, crdClient, kubeClient, dedicatedNameSpace)
		deploymentGridController.FedDeploymentGridController = federationdeploymentGridController

		federationsvcGridController = service.NewFedServiceGridController(
			controllerConfig.ServiceGridInformer, controllerConfig.FedServiceGridInformer, controllerConfig.NodeInformer, fedCrdClient, crdClient, kubeClient, dedicatedNameSpace)
		serviceGridController.FedServiceGridController = federationsvcGridController
	}

	controllerConfig.Run(ctx.Done())
	go deploymentGridController.Run(workerNum, ctx.Done())
	go statefulSetGridController.Run(workerNum, ctx.Done())
	go serviceGridController.Run(workerNum, ctx.Done())
	if fedCrdClient != nil {
		go federationdeploymentGridController.Run(workerNum, ctx.Done())
		go federationsvcGridController.Run(workerNum, ctx.Done())
	}
	<-ctx.Done()
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}

func generateParentDedicatedKubeConfig(kubeClient *clientset.Clientset) (*restclient.Config, string, error) {
	secret, err := kubeClient.CoreV1().Secrets(common.ParentClusterSecretNamespace).Get(context.TODO(),
		common.ParentClusterSecretName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Errorf("failed to get secretFromParentCluster: %v", err)
			return nil, "", err
		} else {
			klog.Infof("secretFromParentCluster is none")
			return nil, "", nil
		}
	} else {
		parentDedicatedKubeConfig, err := controllerutil.GenerateKubeConfigFromToken(string(secret.Data[common.ParentURLKey]),
			string(secret.Data[v1.ServiceAccountTokenKey]), secret.Data[v1.ServiceAccountRootCAKey], 2)
		if err != nil {
			klog.Errorf("failed to generate parentDedicatedKubeConfig: %v", err)
			return nil, "", err
		}
		return parentDedicatedKubeConfig, string(secret.Data["namespace"]), nil
	}
}

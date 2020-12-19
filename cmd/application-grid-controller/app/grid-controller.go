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
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"os"
	"superedge/cmd/application-grid-controller/app/options"
	"superedge/pkg/application-grid-controller/config"
	"superedge/pkg/application-grid-controller/controller/deployment"
	"superedge/pkg/application-grid-controller/controller/service"
	"superedge/pkg/application-grid-controller/prepare"
	"superedge/pkg/util"
	"superedge/pkg/version"
	"superedge/pkg/version/verflag"
	"time"

	crdClientset "superedge/pkg/application-grid-controller/generated/clientset/versioned"
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

			var electionChecker *leaderelection.HealthzAdaptor
			if !o.LeaderElect {
				runController(context.TODO(), apiextensionClient, kubeClient, crdClient, o.Worker, o.SyncPeriod)
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
						/*
						 */
						runController(ctx, apiextensionClient, kubeClient, crdClient, o.Worker, o.SyncPeriod)
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
	apiextensionClient *apiextclientset.Clientset, kubeClient *clientset.Clientset, crdClient *crdClientset.Clientset,
	workerNum, syncPeriod int) {
	crdP := prepare.NewCRDPreparator(apiextensionClient)

	/* create crd, wait for crd ready
	 */
	if err := crdP.Prepare(schema.GroupVersionKind{
		Group:   "superedge.io",
		Version: "v1",
		Kind:    "DeploymentGrid",
	}, schema.GroupVersionKind{
		Group:   "superedge.io",
		Version: "v1",
		Kind:    "ServiceGrid",
	}); err != nil {
		klog.Fatalf("can't create crds: %v", err)
	}

	controllerConfig := config.NewControllerConfig(crdClient, kubeClient, time.Second*time.Duration(syncPeriod))
	deploymentGridController := deployment.NewDeploymentGridController(
		controllerConfig.DeploymentGridInformer, controllerConfig.DeploymentInformer, controllerConfig.NodeInformer,
		kubeClient, crdClient)
	serviceGridController := service.NewServiceGridController(controllerConfig.ServiceGridInformer, controllerConfig.ServiceInformer,
		kubeClient, crdClient)

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	controllerConfig.Run(ctx.Done())
	go deploymentGridController.Run(workerNum, ctx.Done())
	go serviceGridController.Run(workerNum, ctx.Done())
	<-ctx.Done()
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}

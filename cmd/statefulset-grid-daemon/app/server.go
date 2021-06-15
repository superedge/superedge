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
	"github.com/superedge/superedge/cmd/statefulset-grid-daemon/app/options"
	crdClientset "github.com/superedge/superedge/pkg/application-grid-controller/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/config"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/controller"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/hosts"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientset "k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"os"
	"time"
)

func NewStatefulSetGridDaemonCommand() *cobra.Command {
	o := options.NewStatefulsetGridDaemonOptions()

	cmd := &cobra.Command{
		Use: "statefulset-grid-daemon",
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

			hosts := hosts.NewHosts(o.HostPath)
			if _, err := hosts.LoadHosts(); err != nil {
				klog.Fatalf("init load hosts file err: %v", err)
			}

			var electionChecker *leaderelection.HealthzAdaptor
			if !o.LeaderElect {
				runController(context.TODO(), kubeClient, crdClient, o.Worker, o.SyncPeriod, o.SyncPeriodAsWhole, o.HostName, hosts)
				panic("unreachable")
			}

			electionChecker = leaderelection.NewLeaderHealthzAdaptor(time.Second * 20)

			id, err := os.Hostname()
			if err != nil {
				klog.Fatalf("failed to get hostname %v", err)
			}
			id = id + "_" + string(uuid.NewUUID())
			eventRecorder := createRecorder(kubeClient, options.StatefulSetGridDaemonUserAgent)
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
						runController(ctx, kubeClient, crdClient, o.Worker, o.SyncPeriod, o.SyncPeriodAsWhole, o.HostName, hosts)
					},
					OnStoppedLeading: func() {
						klog.Fatalf("leaderelection lost")
					},
				},
				WatchDog: electionChecker,
				Name:     options.StatefulSetGridDaemonUserAgent,
			})
			panic("unreachable")
		},
	}

	fs := cmd.Flags()
	o.AddFlags(fs)

	return cmd
}

func runController(parent context.Context,
	kubeClient *clientset.Clientset, crdClient *crdClientset.Clientset,
	workerNum, syncPeriod, syncPeriodAsWhole int, hostName string, hosts *hosts.Hosts) {

	controllerConfig := config.NewControllerConfig(kubeClient, crdClient, time.Second*time.Duration(syncPeriod))

	statefulSetGridDaemonController := controller.NewStatefulSetGridDaemonController(
		controllerConfig.NodeInformer, controllerConfig.PodInformer, controllerConfig.StatefulSetInformer,
		controllerConfig.StatefulSetGridInformer, controllerConfig.ServiceInformer, kubeClient, hostName, hosts)

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	controllerConfig.Run(ctx.Done())
	go statefulSetGridDaemonController.Run(workerNum, syncPeriodAsWhole, ctx.Done())
	<-ctx.Done()
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}

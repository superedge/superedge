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
	"github.com/superedge/superedge/cmd/apps-manager/app/options"
	"github.com/superedge/superedge/pkg/statefulset-grid-daemon/hosts"

	"github.com/superedge/superedge/pkg/apps-manager/config"
	"github.com/superedge/superedge/pkg/apps-manager/controller"
	crdClientset "github.com/superedge/superedge/pkg/apps-manager/generated/clientset/versioned"
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

func NewAppsManagerDaemonCommand() *cobra.Command {
	siteOptions := options.NewAppsManagerDaemonOptions()
	cmd := &cobra.Command{
		Use: "apps-manager",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			kubeconfig, err := clientcmd.BuildConfigFromFlags(siteOptions.Master, siteOptions.Kubeconfig)
			if err != nil {
				klog.Fatalf("failed to create kubeconfig: %v", err)
			}

			kubeconfig.QPS = siteOptions.QPS
			kubeconfig.Burst = siteOptions.Burst
			kubeClient := clientset.NewForConfigOrDie(kubeconfig)
			crdClient := crdClientset.NewForConfigOrDie(kubeconfig)

			hosts := hosts.NewHosts(siteOptions.HostPath)
			//if _, err := hosts.LoadHosts(); err != nil {
			//	klog.Fatalf("init load hosts file err: %v", err)
			//}

			// not leade elect
			if !siteOptions.LeaderElect { // todo: 不能再选主，每个node都只处理自己的node的任务
				runController(context.TODO(), kubeClient, crdClient, siteOptions.Worker, siteOptions.SyncPeriod, siteOptions.SyncPeriodAsWhole, siteOptions.HostName, hosts)
				panic("unreachable")
			}

			hostname, err := os.Hostname()
			if err != nil {
				klog.Fatalf("failed to get hostname %v", err)
			}
			identityId := hostname + "_" + string(uuid.NewUUID())

			// Create resource lock
			copyConfig := *kubeconfig
			copyConfig.Timeout = time.Second * siteOptions.RenewDeadline.Duration
			leaderElectionClient := clientset.NewForConfigOrDie(restclient.AddUserAgent(&copyConfig, "leader-election"))
			resourceLock, err := resourcelock.New(siteOptions.ResourceLock, siteOptions.ResourceNamespace, siteOptions.ResourceName,
				leaderElectionClient.CoreV1(),
				leaderElectionClient.CoordinationV1(),
				resourcelock.ResourceLockConfig{
					Identity:      identityId,
					EventRecorder: createRecorder(kubeClient, options.SiteManagerDaemonUserAgent),
				})
			if err != nil {
				klog.Fatalf("error creating lock: %v", err)
			}

			// leader running controller
			var electionChecker *leaderelection.HealthzAdaptor
			electionChecker = leaderelection.NewLeaderHealthzAdaptor(time.Second * 20)
			leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
				Lock:          resourceLock,
				LeaseDuration: siteOptions.LeaseDuration.Duration,
				RenewDeadline: siteOptions.RenewDeadline.Duration,
				RetryPeriod:   siteOptions.RetryPeriod.Duration,
				Callbacks: leaderelection.LeaderCallbacks{
					OnStartedLeading: func(ctx context.Context) {
						runController(ctx, kubeClient, crdClient, siteOptions.Worker, siteOptions.SyncPeriod, siteOptions.SyncPeriodAsWhole, siteOptions.HostName, hosts)
					},
					OnStoppedLeading: func() {
						klog.Fatalf("leaderelection lost")
					},
				},
				WatchDog: electionChecker,
				Name:     options.SiteManagerDaemonUserAgent,
			})
			panic("unreachable")
		},
	}

	fs := cmd.Flags()
	siteOptions.AddFlags(fs)

	return cmd
}

func runController(parent context.Context,
	kubeClient *clientset.Clientset, crdClient *crdClientset.Clientset,
	workerNum, syncPeriod, syncPeriodAsWhole int, hostName string, hosts *hosts.Hosts) {

	controllerConfig := config.NewControllerConfig(kubeClient, crdClient, time.Second*time.Duration(syncPeriod))

	appsManagerDaemonController := controller.NewAppsManagerDaemonController(
		controllerConfig.NodeInformer, controllerConfig.PodInformer, controllerConfig.EDeployInformer,
		controllerConfig.ServiceInformer, kubeClient, crdClient, hostName, hosts)

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	controllerConfig.Run(ctx.Done())
	go appsManagerDaemonController.Run(workerNum, syncPeriodAsWhole, ctx.Done())
	<-ctx.Done()
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}

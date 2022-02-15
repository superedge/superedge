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
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	clientgokubescheme "k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/cmd/site-manager/app/options"
	"github.com/superedge/superedge/pkg/site-manager/config"
	"github.com/superedge/superedge/pkg/site-manager/constant"
	"github.com/superedge/superedge/pkg/site-manager/controller"
	crdclientset "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned"
	"github.com/superedge/superedge/pkg/site-manager/utils"
	"github.com/superedge/superedge/pkg/util"
	utilkubeclient "github.com/superedge/superedge/pkg/util/kubeclient"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
)

func NewSiteManagerDaemonCommand() *cobra.Command {
	siteOptions := options.NewSiteManagerDaemonOptions()
	cmd := &cobra.Command{
		Use: "site-manager",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Site-manager Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			kubeconfig, err := clientcmd.BuildConfigFromFlags(siteOptions.Master, siteOptions.Kubeconfig)
			if err != nil {
				klog.Fatalf("Failed to create kubconfig: %#v", err)
			}

			kubeconfig.QPS = siteOptions.QPS
			kubeconfig.Burst = siteOptions.Burst
			kubeClient := clientset.NewForConfigOrDie(kubeconfig)
			crdClient := crdclientset.NewForConfigOrDie(kubeconfig)
			extensionsClient, err := apiextensionsclient.NewForConfig(kubeconfig)
			if err != nil {
				klog.Fatalf("Error instantiating apiextensions client: %s", err.Error())
			}

			runConfig := func(ctx context.Context) {
				stop := ctx.Done()
				if siteOptions.EnsureCrd {
					wait.PollImmediateUntil(time.Second*5, func() (bool, error) {
						utilkubeclient.CreateOrUpdateCustomResourceDefinition(extensionsClient, constant.CRDNodeUnitDefinitionYaml, "")
						utilkubeclient.CreateOrUpdateCustomResourceDefinition(extensionsClient, constant.CRDNodegroupDefinitionYaml, "")
						time.Sleep(5 * time.Second)
						if err := utils.CreateDefaultUnit(crdClient); err != nil {
							klog.Errorf("Create default unit error: %#v", err)
						}
						return true, nil
					}, stop)
				}
			}
			go runConfig(context.TODO())

			// not leade elect
			if !siteOptions.LeaderElect {
				runController(context.TODO(), kubeClient, crdClient, siteOptions.Worker, siteOptions.SyncPeriod, siteOptions.SyncPeriodAsWhole)
				panic("Start site-manager failed\n")
			}

			hostname, err := os.Hostname()
			if err != nil {
				klog.Fatalf("Failed to get hostname %#v", err)
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
				klog.Fatalf("Creating leader elect lock error %#v", err)
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
						runController(ctx, kubeClient, crdClient, siteOptions.Worker, siteOptions.SyncPeriod, siteOptions.SyncPeriodAsWhole)
					},
					OnStoppedLeading: func() {
						klog.Fatalf("Leader election lost")
					},
				},
				WatchDog: electionChecker,
				Name:     options.SiteManagerDaemonUserAgent,
			})
			panic("Start site-manager failed\n")
		},
	}

	fs := cmd.Flags()
	siteOptions.AddFlags(fs)

	return cmd
}

func runController(parent context.Context, kubeClient *clientset.Clientset,
	crdClient *crdclientset.Clientset, workerNum, syncPeriod, syncPeriodAsWhole int) {

	controllerConfig := config.NewControllerConfig(kubeClient, crdClient, time.Second*time.Duration(syncPeriod))
	sitesManagerDaemonController := controller.NewSitesManagerDaemonController(controllerConfig.NodeInformer,
		controllerConfig.NodeUnitInformer, controllerConfig.NodeGroupInformer, kubeClient, crdClient)

	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	controllerConfig.Run(ctx.Done())
	go sitesManagerDaemonController.Run(workerNum, syncPeriodAsWhole, ctx.Done())
	<-ctx.Done()
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(clientgokubescheme.Scheme, v1.EventSource{Component: userAgent})
}

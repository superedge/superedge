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
	"github.com/superedge/superedge/cmd/penetrator-controller/app/options"
	clientset "github.com/superedge/superedge/pkg/penetrator/client/clientset/versioned"
	"github.com/superedge/superedge/pkg/penetrator/client/clientset/versioned/scheme"
	"github.com/superedge/superedge/pkg/penetrator/client/informers/externalversions"
	"github.com/superedge/superedge/pkg/penetrator/constants"
	"github.com/superedge/superedge/pkg/penetrator/operator/admission"
	ntctx "github.com/superedge/superedge/pkg/penetrator/operator/context"
	"github.com/superedge/superedge/pkg/penetrator/operator/controller"
	"github.com/superedge/superedge/pkg/penetrator/operator/crd"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeclient"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"runtime"
	"time"
)

func NewOperatorCommand() *cobra.Command {
	o := options.NewOperatorOptions()

	cmd := &cobra.Command{
		Use: "AddNodeCOntrolller",
		Run: func(cmd *cobra.Command, args []string) {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			cfg, err := kubeclient.GetKubeConfig(o.KubeConfig)
			if err != nil {
				klog.Fatalf("failed to build in-cluster kubeconfig, error: %v", err)
			}

			cfg.QPS = float32(o.QPS)
			cfg.Burst = o.Burst

			ctx := ntctx.NewContext(context.Background())

			extensionsClient, err := apiextensionsclient.NewForConfig(cfg)
			if err != nil {
				klog.Fatalf("failed to build apiextensions clientset, error:%v", err)
			}

			err = crd.InstallWithMaxRetry(ctx, extensionsClient, &crd.NodesTaskCustomResourceDefinition, 10)
			if err != nil {
				klog.Fatalf("failed to install Task CustomResourceDefinition, error: %v", err)
			}

			//kubeclient
			kubeClient, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				klog.Fatalf("failed to  build kubernetes clientset: %v", err)
			}

			//nodestaskkubeclient
			nodestaskClient, err := clientset.NewForConfig(cfg)
			if err != nil {
				klog.Fatalf("failed to  build nodestask clientset: %v", err)
			}

			//rlkubeclient
			rlkubeclient, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				klog.Fatalf("failed to build rl kubernetes clientset, error: %v", err)
			}

			//eventkubeclient
			eventKubeclient, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				klog.Fatalf("failed to build event kubernetes, error: %v", err)
			}

			//event recorder
			klog.V(4).Info("creating nodestask event broadcaster")
			eventBroadcaster := record.NewBroadcaster()
			eventBroadcaster.StartLogging(klog.Infof)
			eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: eventKubeclient.CoreV1().Events("")})
			err = scheme.AddToScheme(kubescheme.Scheme)
			if err != nil {
				klog.Fatalf("failed to add penetrator scheme, error: %v", err)
			}
			record := eventBroadcaster.NewRecorder(kubescheme.Scheme, corev1.EventSource{Component: "penetrator-controller"})

			// webhook
			if o.EnableAdmissionControl {
				if o.AdmissionControlServerCert == "" {
					klog.Fatal("admission-control-server-cert must be given")
				}
				if o.AdmissionControlServerKey == "" {
					klog.Fatal("admission-control-server-key must be given")
				}

				server := &http.Server{
					Addr:    o.AdmissionControlListenAddr,
					Handler: admission.Handler(kubeClient, nodestaskClient, ctx),
				}
				go func() {
					if err := server.ListenAndServeTLS(o.AdmissionControlServerCert, o.AdmissionControlServerKey); err != nil {
						klog.Fatal(err)
					}
				}()
			}

			//nodestaskcontroller
			run := func(controllerctx context.Context) {
				nodestaskInformerFactory := externalversions.NewSharedInformerFactory(nodestaskClient, time.Second*30)
				nodestaskController := controller.NewNodeTaskController(kubeClient, nodestaskClient, nodestaskInformerFactory.Nodestask().V1beta1().NodeTasks(), ctx, record)

				nodestaskInformerFactory.Start(make(<-chan struct{}, 0))
				nodestaskController.Run(runtime.NumCPU(), controllerctx.Done())
			}

			//leader election
			id, err := os.Hostname()
			if err != nil {
				klog.Fatalf("failed to get hostname, error: %v", err)
			}
			if os.Getenv(constants.EnvOperatorPodName) != "" {
				id = id + "-" + os.Getenv(constants.EnvOperatorPodName)
			}
			rl, err := resourcelock.New(resourcelock.EndpointsResourceLock, os.Getenv(constants.EnvOperatorNamespace), "penetrator-controller", rlkubeclient.CoreV1(), rlkubeclient.CoordinationV1(), resourcelock.ResourceLockConfig{
				Identity:      id,
				EventRecorder: record,
			})
			if err != nil {
				klog.Fatalf("failed to create lock, error:%v", err)
			}

			leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
				Lock:          rl,
				LeaseDuration: o.LeaseDuration,
				RenewDeadline: o.RenewDeadline,
				RetryPeriod:   o.RetryPeriod,
				Callbacks: leaderelection.LeaderCallbacks{
					OnStartedLeading: run,
					OnStoppedLeading: func() {
						klog.Fatalf("leader election lost ")
					},
				},
			})

		},
	}
	fs := cmd.Flags()
	o.AddFlags(fs)

	return cmd
}

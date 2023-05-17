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

package options

import (
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config"
)

const (
	SiteManagerDaemonUserAgent = "site-manager"
)

type Options struct {
	Burst             int
	SyncPeriod        int
	SyncPeriodAsWhole int
	Worker            int
	QPS               float32
	Master            string
	Kubeconfig        string
	EnsureCrd         bool
	FeatureGates      map[string]bool
	config.LeaderElectionConfiguration
}

func NewSiteManagerDaemonOptions() *Options {
	featureGates := make(map[string]bool)
	featureGates["ServiceTopology"] = true
	featureGates["EndpointSlice"] = true

	return &Options{
		SyncPeriod:        30,
		SyncPeriodAsWhole: 30,
		Burst:             1000,
		QPS:               float32(1000),
		Worker:            runtime.NumCPU(),
		LeaderElectionConfiguration: config.LeaderElectionConfiguration{
			ResourceLock:      "site-manager",
			ResourceNamespace: metav1.NamespaceSystem,
			ResourceName:      SiteManagerDaemonUserAgent,
			RetryPeriod:       metav1.Duration{Duration: time.Second * time.Duration(2)},
			LeaseDuration:     metav1.Duration{Duration: time.Second * time.Duration(15)},
			RenewDeadline:     metav1.Duration{Duration: time.Second * time.Duration(10)},
		},
		FeatureGates: featureGates,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Master, "master", o.Master, "apiserver master address")
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "kubeconfig path, empty path means in cluster mode")
	fs.BoolVar(&o.EnsureCrd, "ensure-crd", true, "auto apply all crd definition")

	fs.Float32Var(&o.QPS, "kube-qps", o.QPS, "kubeconfig qps setting")
	fs.IntVar(&o.Burst, "kube-burst", o.Burst, "kubeconfig burst setting")
	fs.Var(cliflag.NewMapStringBool(&o.FeatureGates), "feature-gates", "A set of key=value pairs that describe feature gates for alpha/experimental features. "+
		"Options are:\n"+strings.Join(utilfeature.DefaultMutableFeatureGate.KnownFeatures(), "\n"))
	fs.IntVar(&o.SyncPeriod, "sync-period", o.SyncPeriod, "Period for syncing the objects")
	fs.IntVar(&o.SyncPeriodAsWhole, "sync-period-as-whole", o.SyncPeriodAsWhole, "Period for syncing the dns hosts as whole")
	fs.IntVar(&o.Worker, "worker", o.Worker, "worker number of controller")

	fs.BoolVar(&o.LeaderElect, "leader-elect", o.LeaderElect, ""+
		"Start a leader election client and gain leadership before "+
		"executing the main loop. Enable this when running replicated "+
		"components for high availability.")
	fs.DurationVar(&o.LeaseDuration.Duration, "leader-elect-lease-duration", o.LeaseDuration.Duration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&o.RenewDeadline.Duration, "leader-elect-renew-deadline", o.RenewDeadline.Duration, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&o.RetryPeriod.Duration, "leader-elect-retry-period", o.RetryPeriod.Duration, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
	fs.StringVar(&o.ResourceLock, "leader-elect-resource-lock", o.ResourceLock, ""+
		"The type of resource object that is used for locking during "+
		"leader election. Supported options are `endpoints` (default) and `configmaps`.")
	fs.StringVar(&o.ResourceName, "leader-elect-resource-name", o.ResourceName, ""+
		"The name of resource object that is used for locking during "+
		"leader election.")
	fs.StringVar(&o.ResourceNamespace, "leader-elect-resource-namespace", o.ResourceNamespace, ""+
		"The namespace of resource object that is used for locking during "+
		"leader election.")
}

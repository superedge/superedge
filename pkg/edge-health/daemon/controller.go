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

package daemon

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"time"
)

func (ehd *EdgeHealthDaemon) Prepare(stopCh <-chan struct{}) {
	if !cache.WaitForNamedCacheSync("edge-health", stopCh,
		ehd.cfg.NodeInformer.Informer().HasSynced, ehd.cfg.ConfigMapInformer.Informer().HasSynced) {
		return
	}
}

func (ehd *EdgeHealthDaemon) Check(stopCh <-chan struct{}) {
	go wait.Until(ehd.SyncNodeList, time.Duration(ehd.cfg.Check.HealthCheckPeriod)*time.Second, stopCh)
	go wait.Until(ehd.ExecuteCheck, time.Duration(ehd.cfg.Check.HealthCheckPeriod)*time.Second, stopCh)
}

func (ehd *EdgeHealthDaemon) PrepareAndCheck(stopCh <-chan struct{}) {
	ehd.Prepare(stopCh)
	ehd.Check(stopCh)
}

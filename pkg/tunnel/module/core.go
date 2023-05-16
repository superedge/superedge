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

package module

import (
	"github.com/superedge/superedge/pkg/tunnel/tunnelcontext"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
)

func LoadModules(mode string) {
	modules := GetModules()
	for n, m := range modules {
		tunnelcontext.GetContext().AddModule(n)
		klog.Infof("starting module:%s", m.Name())
		m.Start(mode)
		klog.Infof("start module:%s success !", m.Name())
	}

}

func ShutDown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	s := <-c
	klog.Info("got os signal " + s.String())
	modules := GetModules()
	for name, module := range modules {
		klog.Info("cleanup module " + name)
		module.CleanUp()
	}
}

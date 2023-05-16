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

package tunnelcontext

import (
	"flag"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"os"
	"testing"

	"k8s.io/klog/v2"
)

func Test_Send_Msg(t *testing.T) {
	fs := flag.NewFlagSet("Test_Send_Msg", flag.ExitOnError)
	klog.InitFlags(fs)
	err := fs.Set("v", "8")
	if err != nil {
		t.Errorf("failed to set klog level err: %s", err)
		return
	}
	fs.Parse(os.Args[0:])
	GetContext().AddNode("node1")

}

func Test_Klog(t *testing.T) {
	fs := flag.NewFlagSet("Test_Klog", flag.ExitOnError)
	klog.InitFlags(fs)
	err := fs.Set("v", "8")
	if err != nil {
		t.Errorf("failed to set klog level err: %s", err)
		return
	}
	fs.Parse(os.Args[0:])
	GetContext().AddModule(util.STREAM)
}

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

package main

import (
	goflag "flag"
	"k8s.io/klog"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"

	"superedge/cmd/edgeadm/app"
)

const (
	bashCompleteFile = "/etc/bash_completion.d/edgeadm.bash_complete"
)

func main() {
	cmd := app.NewEdgeadmCommand()

	pflag.CommandLine.SetNormalizeFunc(flag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Set("logtostderr", "true")

	klog.InitFlags(nil)
	defer klog.Flush()

	// We do not want these flags to show up in --help
	// These MarkHidden calls must be after the lines above
	pflag.CommandLine.MarkHidden("version")
	pflag.CommandLine.MarkHidden("log-dir")

	cmd.GenBashCompletionFile(bashCompleteFile)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	return
}

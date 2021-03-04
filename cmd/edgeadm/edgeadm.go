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
	"flag"
	"os"
	"path"

	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/cmd/edgeadm/app"
	"github.com/superedge/superedge/pkg/edgeadm/constant"
)

const (
	bashCompleteFile = "/etc/bash_completion.d/edgeadm.bash_complete"
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	klogSet()
	defer klog.Flush()

	cmd := app.NewEdgeadmCommand(os.Stdin, os.Stdout, os.Stderr)
	cmd.GenBashCompletionFile(bashCompleteFile)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	return
}

// We do not want these flags to show up in --help
// These MarkHidden calls must be after the lines above
func klogSet() {
	pflag.CommandLine.MarkHidden("log-dir")
	pflag.CommandLine.MarkHidden("version")
	pflag.CommandLine.MarkHidden("vmodule")
	pflag.CommandLine.MarkHidden("one-output")
	pflag.CommandLine.MarkHidden("logtostderr")
	pflag.CommandLine.MarkHidden("skip-headers")
	pflag.CommandLine.MarkHidden("add-dir-header")
	pflag.CommandLine.MarkHidden("alsologtostderr")
	pflag.CommandLine.MarkHidden("stderrthreshold")
	pflag.CommandLine.MarkHidden("log-backtrace-at")
	pflag.CommandLine.MarkHidden("skip-log-headers")
	pflag.CommandLine.MarkHidden("log-file-max-size")
	pflag.CommandLine.MarkHidden("log-flush-frequency")

	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "true")
	os.MkdirAll(path.Dir(constant.EdgeClusterLogFile), 0755)
	pflag.Set("log_file", constant.EdgeClusterLogFile)
	flag.Parse()
}

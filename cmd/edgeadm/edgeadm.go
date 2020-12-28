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
	"fmt"
	"k8s.io/klog"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"

	"github.com/superedge/superedge/cmd/edgeadm/app"
)

const (
	bashCompleteFile = "/etc/bash_completion.d/edgeadm.bash_complete"
)

func main() {
	cmd := app.NewEdgeadmCommand()

	klog.InitFlags(goflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(flag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	if err := pflag.Set("logtostderr", "true"); err != nil {
		fmt.Printf("Failed to set logtostderr, %s\n", err)
		os.Exit(1)
	}

	defer klog.Flush()

	// We do not want these flags to show up in --help
	// These MarkHidden calls must be after the lines above
	if err := pflag.CommandLine.MarkHidden("version"); err != nil {
		fmt.Printf("Set CommandLine MarkHidden version failed, %s\n", err)
		os.Exit(1)
	}
	if err := pflag.CommandLine.MarkHidden("log-dir"); err != nil {
		fmt.Printf("Set CommandLine MarkHidden log-dir failed, %s\n", err)
		os.Exit(1)
	}

	cmd.GenBashCompletionFile(bashCompleteFile) // nolint
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	return // nolint
}

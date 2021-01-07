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
	"github.com/spf13/pflag"
	"github.com/superedge/superedge/cmd/edge-health/app"
	"github.com/superedge/superedge/pkg/edge-health/util"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"os"
)

func main() {
	ctx, _ := util.SignalWatch()
	command := app.NewEdgeHealthCommand(ctx)

	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

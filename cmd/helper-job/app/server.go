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
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/helper-job/deploy"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
)

func NewHelperCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "helper",
		Long:         `helper is a job about deploy init edge node`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()
			klog.Infof("Versions: %#v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			// complete all default server options
			if err := Complete(); err != nil {
				return err
			}

			// validate options
			if err := Validate(); err != nil {
				return err
			}

			return deploy.Run()
		},
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

func Complete() error {
	return nil
}

func Validate() error {
	return nil
}

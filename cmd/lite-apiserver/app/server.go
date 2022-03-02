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
	"fmt"
	"io"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"github.com/moby/term"
	"github.com/spf13/cobra"

	"github.com/superedge/superedge/cmd/lite-apiserver/app/options"
	"github.com/superedge/superedge/pkg/lite-apiserver/server"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/version"
	"github.com/superedge/superedge/pkg/version/verflag"
)

var ComponentName = "lite-apiserver"

func NewServerCommand() *cobra.Command {
	o := options.NewServerRunOptions()
	cmd := &cobra.Command{
		Use:  ComponentName,
		Long: `The lite-apiserver is the daemon for node, proxy all request for kube-apiserver. And cache all get request body`,
		// stop printing usage when the command errors
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()

			klog.Infof("Versions: %+v\n", version.Get())
			util.PrintFlags(cmd.Flags())

			// complete all default server options
			if err := o.Complete(); err != nil {
				klog.Errorf("complete options error: %v", err)
				return err
			}

			// validate options
			if errs := o.Validate(); len(errs) != 0 {
				klog.Errorf("validate options error: %v", errs)
				return utilerrors.NewAggregate(errs)
			}

			return Run(o, util.SetupSignalHandler())
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	fs := cmd.Flags()
	namedFlagSets := o.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := TerminalSize(cmd.OutOrStdout())
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	return cmd
}

// Run runs the specified APIServer.  This should never exit.
func Run(serverOptions *options.ServerRunOptions, stopCh <-chan struct{}) error {

	server, err := server.CreateServer(serverOptions, stopCh)
	if err != nil {
		klog.Errorf("create lite-apiserver error: %v", err)
		return err
	}

	return server.Run()
}

// TerminalSize returns the current width and height of the user's terminal. If it isn't a terminal,
// nil is returned. On error, zero values are returned for width and height.
// Usually w must be the stdout of the process. Stderr won't work.
func TerminalSize(w io.Writer) (int, int, error) {
	outFd, isTerminal := term.GetFdInfo(w)
	if !isTerminal {
		return 0, 0, fmt.Errorf("given writer is no terminal")
	}
	winsize, err := term.GetWinsize(outFd)
	if err != nil {
		klog.Errorf("get window size error: %v", err)
		return 0, 0, err
	}
	return int(winsize.Width), int(winsize.Height), nil
}

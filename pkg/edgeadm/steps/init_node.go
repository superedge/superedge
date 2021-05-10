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

package steps

import (
	"fmt"

	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
)

var (
	initNodeLongDesc = cmdutil.LongDesc(`Init node before install node or master of Kubernetes.`)
)

// Init node before install node or master of Kubernetes
func NewInitNodePhase() workflow.Phase {
	return workflow.Phase{
		Name:  "init-node",
		Short: "Init node before install node or master of Kubernetes",
		Long:  initNodeLongDesc,
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          initNodeLongDesc,
				InheritFlags:   getInitNodePhaseFlags("all"),
				RunAllSiblings: true,
			},
			{
				Name:         "off-swap",
				Short:        "Off swap of init Kubernetes node",
				InheritFlags: getInitNodePhaseFlags("off-swap"),
				Run:          runOffSwap,
			},
			{
				Name:         "stop-firewall",
				Short:        "Stop firewall of init Kubernetes node",
				InheritFlags: getInitNodePhaseFlags("stop-firewall"),
				Run:          stopFirewall,
			},
			{
				Name:         "set-sysctl",
				Short:        "Set system parameters for Kubernetes nod by sysctl tools",
				InheritFlags: getInitNodePhaseFlags("set-sysctl"),
				Run:          setSysctl,
			},
			{
				Name:         "load-kernel",
				Short:        "Set kernel modules of init Kubernetes node",
				InheritFlags: getInitNodePhaseFlags("load-kernel"),
				Run:          loadKernelModule,
			},
		},
	}
}

// Init node flags
func getInitNodePhaseFlags(name string) []string {
	flags := []string{
		options.KubeconfigPath,
	}
	if name == "all" || name == "off-swap" {
	}
	if name == "all" || name == "set-sysctl" {
	}
	if name == "all" || name == "load-kernel" {
	}
	if name == "all" || name == "set-hostname" {
	}
	if name == "all" || name == "stop-firewall" {
	}
	return flags
}

// set hostname about node or master of Kubernetes
func setHostname(c workflow.RunData) error {
	loadIP, err := util.GetLocalIP()
	if err != nil {
		return err
	}
	steHostname := fmt.Sprintf("hostnamectl set-hostname node-%s", loadIP)
	if _, _, err := util.RunLinuxCommand(steHostname); err != nil {
		return err
	}
	return err
}

// set off swap
func runOffSwap(c workflow.RunData) error {
	if _, _, err := util.RunLinuxCommand(constant.SwapOff); err != nil {
		return err
	}
	return nil
}

// stop firewalld
func stopFirewall(c workflow.RunData) error {
	if _, _, err := util.RunLinuxCommand(constant.StopFireWall); err != nil {
		return err
	}
	return nil
}

// set system parameters by sysctl
func setSysctl(c workflow.RunData) error {
	setNetIPv4 := util.SetFileContent(constant.SysctlFile, "^net.ipv4.ip_forward.*", "net.ipv4.ip_forward = 1")
	if _, _, err := util.RunLinuxCommand(setNetIPv4); err != nil {
		return err
	}
	setNetBridge := util.SetFileContent(constant.SysctlFile, "^net.bridge.bridge-nf-call-iptables.*", "net.bridge.bridge-nf-call-iptables = 1")
	if _, _, err := util.RunLinuxCommand(setNetBridge); err != nil {
		return err
	}

	setSysctl := fmt.Sprintf("cat <<EOF >%s \n%s\nEOF", constant.SysctlK8sConf, constant.SysConf)
	if _, _, err := util.RunLinuxCommand(setSysctl); err != nil {
		return err
	}

	loadIPtables := fmt.Sprintf("sysctl --system")
	if _, _, err := util.RunLinuxCommand(loadIPtables); err != nil {
		return err
	}
	return nil
}

// load kernel module require of install Kubernetes
func loadKernelModule(c workflow.RunData) error {
	modules := []string{
		"ip_vs",
		"ip_vs_sh",
		"ip_vs_rr",
		"ip_vs_wrr",
		"iptable_nat",
		"nf_conntrack_ipv4",
	}
	if _, _, err := util.RunLinuxCommand("modinfo br_netfilter"); err == nil {
		modules = append(modules, "br_netfilter")
	}

	kernelModule := ""
	for _, module := range modules {
		kernelModule += fmt.Sprintf("modprobe -- %s\n", module)
	}

	setKernelModule := fmt.Sprintf("cat <<EOF >%s \n%s\nEOF", constant.IPvsModulesFile, kernelModule)
	if _, _, err := util.RunLinuxCommand(setKernelModule); err != nil {
		return err
	}

	if _, _, err := util.RunLinuxCommand(constant.KernelModule); err != nil {
		return err
	}

	return nil
}

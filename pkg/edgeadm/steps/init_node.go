package steps

import (
	"fmt"

	"github.com/superedge/superedge/pkg/edgeadm/constant"
	"github.com/superedge/superedge/pkg/util"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/options"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/phases/workflow"
	cmdutil "github.com/superedge/superedge/pkg/util/kubeadm/app/cmd/util"
)

var (
	initNodeLongDesc = cmdutil.LongDesc(`Init node before install.`)
)

// NewAddonPhase returns the addon Cobra command
func NewInitNodePhase() workflow.Phase { //todo 独立成一个独立的函数，为后面实现kube-*集群master无缝部署，加简便的添加边缘节点
	return workflow.Phase{
		Name:  "init-node",
		Short: "Init node",
		Long:  initNodeLongDesc, //todo
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          "Init node before install",
				InheritFlags:   getInitNodePhaseFlags("all"),
				RunAllSiblings: true,
			},
			{
				Name:         "set-hostname",
				Short:        "Set hostname of init edge Kubernetes cluster",
				Long:         kubeProxyAddonLongDesc, //todo
				InheritFlags: getInitNodePhaseFlags("set-hostname"),
				Run:          setHostname,
			},
			{
				Name:         "off-swap",
				Short:        "Off swap of init edge Kubernetes cluster",
				Long:         kubeProxyAddonLongDesc, //todo
				InheritFlags: getInitNodePhaseFlags("off-swap"),
				Run:          runOffSwap,
			},
			{
				Name:         "stop-firewall",
				Short:        "Stop firewall of init edge Kubernetes cluster",
				Long:         kubeProxyAddonLongDesc, //todo
				InheritFlags: getInitNodePhaseFlags("stop-firewall"),
				Run:          stopFirewall,
			},
			{
				Name:         "set-sysctl",
				Short:        "Set sysctl of init edge Kubernetes cluster",
				Long:         kubeProxyAddonLongDesc, //todo
				InheritFlags: getInitNodePhaseFlags("set-sysctl"),
				Run:          setSysctl,
			},
			{
				Name:         "load-kernel",
				Short:        "Load kernel modules of init edge Kubernetes cluster",
				Long:         kubeProxyAddonLongDesc, //todo
				InheritFlags: getInitNodePhaseFlags("load-kernel"),
				Run:          loadKernelModule,
			},
		},
	}
}

func getInitNodePhaseFlags(name string) []string {
	flags := []string{
		constant.ManifestsDir,
		options.KubeconfigPath,
	}
	if name == "all" || name == "set-hostname" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	if name == "all" || name == "off-swap" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	if name == "all" || name == "stop-firewall" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	if name == "all" || name == "set-sysctl" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}
	if name == "all" || name == "load-kernel" {
		//flags = append(flags,
		//	options.CertificatesDir,
		//)
	}

	return flags
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func setHostname(c workflow.RunData) error {
	loadIP, err := util.GetLocalIP()
	if err != nil {
		return err
	}
	steHostname := fmt.Sprint("hostnamectl set-hostname node-%s", loadIP)
	if _, _, err := util.RunLinuxCommand(steHostname); err != nil {
		return err
	}
	return err
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func runOffSwap(c workflow.RunData) error {
	if _, _, err := util.RunLinuxCommand(constant.SWAP_OFF); err != nil {
		return err
	}
	return nil
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func stopFirewall(c workflow.RunData) error {
	if _, _, err := util.RunLinuxCommand(constant.STOP_FIREWALL); err != nil {
		return err
	}
	return nil
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func setSysctl(c workflow.RunData) error {
	setNetIPv4 := util.SetFileContent(constant.SysctlFile, "^net.ipv4.ip_forward.*", "net.ipv4.ip_forward = 1")
	if _, _, err := util.RunLinuxCommand(setNetIPv4); err != nil {
		return err
	}

	setNetBridge := util.SetFileContent(constant.SysctlFile, "^net.bridge.bridge-nf-call-iptables.*", "net.bridge.bridge-nf-call-iptables = 1")
	if _, _, err := util.RunLinuxCommand(setNetBridge); err != nil {
		return err
	}

	setSysctl := fmt.Sprint("cat <<EOF > /etc/sysctl.d/k8s.conf\n %s \nEOF", constant.SYS_CONF)
	if _, _, err := util.RunLinuxCommand(setSysctl); err != nil {
		return err
	}

	loadIPtables := fmt.Sprint("sysctl --system")
	if _, _, err := util.RunLinuxCommand(loadIPtables); err != nil {
		return err
	}
	return nil
}

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
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

	setKernelModule := fmt.Sprint("cat > /etc/sysconfig/modules/ipvs.modules <<EOF\n %s \nEOF", kernelModule)
	if _, _, err := util.RunLinuxCommand(setKernelModule); err != nil {
		return err
	}

	if _, _, err := util.RunLinuxCommand(constant.KERNEL_MODULE); err != nil {
		return err
	}
	return nil
}

package cmd

type EdgeadmConfig struct {
	IsEnableEdge           bool
	WorkerPath             string
	ManifestsDir           string
	InstallPkgPath         string
	Kubeconfig             string
	TunnelCloudToken       string
	TunnelCoreDNSClusterIP string
}

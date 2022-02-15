package constant

const (
	ISEnableEdge       = "enable-edge"
	DefaultHA          = "default-ha"
	DefaultHAKubeVIP   = "kube-vip"
	InstallPkgPath     = "install-pkg-path"
	ManifestsDir       = "manifests-dir"
	HANetworkInterface = "interface"
	ContainerRuntime   = "runtime"
)

const (
	ControlFormat              = "    "
	InstallPkgPathNote         = "Path of edgeadm kube-* install package"
	InstallPkgNetworkLocation  = ""
	HANetworkDefaultInterface  = "eth0"
	ContainerRuntimeDocker     = "docker"
	ContainerRuntimeContainerd = "containerd"

	DefaultDockerCRISocket     = "/var/run/dockershim.sock"
	DefaultContainerdCRISocket = "/run/containerd/containerd.sock"
)

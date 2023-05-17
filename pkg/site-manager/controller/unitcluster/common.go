package unitcluster

const (
	DefaultKinsNamespace          = "kins-system"
	KinsResourceNameSuffix        = "kins"
	KinsResourceLabelKey          = "site.superedge.io/kins-resource"
	KinsRoleLabelKey              = "site.superedge.io/kins-role"
	KinsUnitClusterClearAnno      = "site.superedge.io/clear-cluster"
	KinsRoleLabelServer           = "server"
	KinsRoleLabelAgent            = "agent"
	DefaultKinsServiceCIDR        = "26.%d.0.%s"
	DefaultKinsNodePortRangeStart = 40000
	DefaultKinsCRIWImage          = "ccr.ccs.tencentyun.com/tkeedge/cri-w:v0.1.0"
	DefaultK3SImage               = "ccr.ccs.tencentyun.com/tkeedge/k3s:v1.22.6-revison-1"

	ParameterK3SImageKey      = "k3s-image"
	ParameterCRIWImageKey     = "criw-image"
	ParameterServiceCIDRKey   = "service-cidr"
	ParameterNodePortRangeKey = "node-port-range"
)

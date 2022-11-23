package unitcluster

const (
	DefaultKinsNamespace     = "kins-system"
	KinsResourceNameSuffix   = "kins"
	KinsResourceLabelKey     = "site.superedge.io/kins-resource"
	KinsRoleLabelKey         = "site.superedge.io/kins-role"
	KinsUnitClusterClearAnno = "site.superedge.io/clear-cluster"
	KinsRoleLabelServer      = "server"
	KinsRoleLabelAgent       = "agent"
	KinsServiceCIDR          = "26.%d.0.%s"
	KinsNodePortRangeStart   = 40000
	KinsCRIWImage            = "ccr.ccs.tencentyun.com/marcest11/marctest:cri-w-4"
	K3SServerImage           = "ccr.ccs.tencentyun.com/marcest11/marctest:k3s-5"
	K3SAgentImage            = "ccr.ccs.tencentyun.com/marcest11/marctest:k3s-5"
)

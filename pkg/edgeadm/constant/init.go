package constant

const (
	EdgeamdDir = "edgeadm-test/"

	DataDir            = EdgeamdDir + "data/"
	EdgeClusterFile    = DataDir + "edgeadm.json"
	EdgeClusterLogFile = DataDir + "edgeadm.log"

	HooksDir             = EdgeamdDir + "hooks/"
	PreInstallHook       = HooksDir + "pre-install"
	PostClusterReadyHook = HooksDir + "post-cluster-ready"
	PostInstallHook      = HooksDir + "post-install"
)

const (
	StatusUnknown = "Unknown"
	StatusDoing   = "Doing"
	StatusSuccess = "Success"
	StatusFailed  = "Failed"
)

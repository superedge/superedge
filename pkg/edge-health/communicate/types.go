package communicate

import "github.com/superedge/superedge/pkg/edge-health/data"

const (
	ProtoTCP      = "tcp"
	ProtoICMP     = "icmp"
	MaxProbeIPs   = 100
	MaxTargetName = 128
)

var ProtoSet = map[string]struct{}{ProtoTCP: {}, ProtoICMP: {}}

type Target struct {
	Name     string `json:"name,omitempty"`
	IP       string `json:"ip"`
	Port     int32  `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Normal   *bool  `json:"normal,omitempty"`
}

type SourceInfo struct {
	SourcePodIP    string `json:"sourcePodIP,omitempty"`
	SourcePodName  string `json:"sourcePodName,omitempty"`
	SourceNodeIP   string `json:"sourceNodeIP,omitempty"`
	SourceNodeName string `json:"sourceNodeName,omitempty"`
}

type Probe struct {
	Targets []*Target `json:"targets,omitempty"`
}

type ProbeResp struct {
	Targets []*Target `json:"targets,omitempty"`
	*SourceInfo
}

type LocalInfoResp struct {
	LocalInfo map[string]map[string]data.ResultDetail `json:"localInfo,omitempty"`
	*SourceInfo
}

package util

import (
	"testing"
)

func TestCloudHttpProxy(t *testing.T) {
	config := NewHttpProxyConfig("-kins.default$:443")
	if config.UseProxy("nj-kins.default:443") {
		t.Error("Domain name matching failed")
	}
}

func TestEdgeHttpProxy(t *testing.T) {
	config := NewHttpProxyConfig("-svc$:443,*.svc.cluster.local:443,*.cluster.local:443")
	if config.UseProxy("k8s-122-svc:443") {
		t.Error("Domain name matching failed")
	}
}

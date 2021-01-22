package util

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestUtil(t *testing.T) {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, "util") {
		wd = filepath.Dir(wd)
	}

	path := path.Join(wd, "host")

	h := NewHosts(path)
	hostsMap, err := h.ParseHosts(h.ReadHostsFile(h.HostPath))
	if err == nil {
		for k, v := range hostsMap {
			t.Logf("%s,%s", k, v)
		}
	}
	h.UpdateHosts(nil, "default", "abc", "svc1")
	t.Log("\n")
	for k, v := range h.HostsMap {
		t.Logf("%s,%s", k, v)
	}
}

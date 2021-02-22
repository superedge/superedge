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

package hosts

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestUtil(t *testing.T) {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, "hosts") {
		wd = filepath.Dir(wd)
	}

	path := path.Join(wd, "hosts")

	h := NewHosts(path)
	hostsMap, err := h.LoadHosts()
	if err == nil {
		for k, v := range hostsMap {
			t.Logf("%s,%s", k, v)
		}
	}
	podDomainInfoToHosts := make(map[string]string)
	// Test unchanged
	podDomainInfoToHosts["statefulsetgrid-demo-0.servicegrid-demo-svc.default.svc.cluster.local"] = "1.2.3.4"
	// Test Update
	podDomainInfoToHosts["statefulsetgrid-demo-1.servicegrid-demo-svc.default.svc.cluster.local"] = "1.2.3.7"
	// Test Add and Delete
	podDomainInfoToHosts["statefulsetgrid-demo-12.servicegrid-demo-svc.default.svc.cluster.local"] = "1.2.3.7"
	h.CheckOrUpdateHosts(podDomainInfoToHosts, "default", "statefulsetgrid-demo", "servicegrid-demo-svc")
	t.Log("\n")
	for k, v := range h.hostsMap {
		t.Logf("%s,%s", k, v)
	}
}

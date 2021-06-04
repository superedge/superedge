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
	"fmt"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var suffix = ".svc.cluster.local"

type Hosts struct {
	hostPath string
	hostsMap map[string]string
	sync.RWMutex
}

func NewHosts(HostPath string) *Hosts {
	return &Hosts{
		hostPath: HostPath,
		hostsMap: make(map[string]string),
	}
}

func (h *Hosts) LoadHosts() (map[string]string, error) {
	h.Lock()
	defer h.Unlock()
	if hostsFileContent, err := ioutil.ReadFile(h.hostPath); err == nil {
		for _, line := range strings.Split(strings.Trim(string(hostsFileContent), " \t\r\n"), "\n") {
			line = strings.Replace(strings.Trim(line, " \t"), "\t", " ", -1)
			if len(line) == 0 || line[0] == ';' || line[0] == '#' {
				continue
			}
			pieces := strings.SplitN(line, " ", 2)
			if len(pieces) != 2 || net.ParseIP(pieces[0]) == nil {
				continue
			}
			if domains := strings.Fields(pieces[1]); len(domains) == 1 {
				h.hostsMap[pieces[1]] = pieces[0]
			}
		}
	} else if os.IsNotExist(err) {
		f, err := os.Create(h.hostPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		h.hostsMap = map[string]string{}
	} else {
		return nil, err
	}
	return h.hostsMap, nil
}

func AppendDomainSuffix(domain, ns string) string {
	return domain + "." + ns + suffix
}

func (h *Hosts) isMatchDomain(domain, ns, ssgName, svcName string) bool {
	match, _ := regexp.MatchString(ssgName+"-"+`[0-9]+`+`\.`+svcName+`\.`+ns+suffix, domain)
	return match
}

func (h *Hosts) CheckOrUpdateHosts(PodDomainInfoToHosts map[string]string, ns, ssgName, svcName string) error {
	h.Lock()
	defer h.Unlock()

	isChanged := false
	for domain, ip := range h.hostsMap {
		// Only cares about those domains that matches statefulset grid headless service pod FQDN records
		if h.isMatchDomain(domain, ns, ssgName, svcName) {
			if curIp, exist := PodDomainInfoToHosts[domain]; !exist {
				// Delete pod relevant domains since it has been deleted
				delete(h.hostsMap, domain)
				klog.V(4).Infof("Deleting dns hosts domain %s and ip %s", domain, ip)
				isChanged = true
			} else if exist && curIp != ip {
				// Update pod relevant domains ip since it has been updated
				h.hostsMap[domain] = curIp
				delete(PodDomainInfoToHosts, domain)
				klog.V(4).Infof("Updating dns hosts domain %s: old ip %s -> ip %s", domain, ip, curIp)
				isChanged = true
			} else if exist && curIp == ip {
				// Stay unchanged
				delete(PodDomainInfoToHosts, domain)
				klog.V(5).Infof("Dns hosts domain %s and ip %s stays unchanged", domain, ip)
			}
		}
	}
	if !isChanged && len(PodDomainInfoToHosts) == 0 {
		// Stay unchanged as a whole
		klog.V(4).Infof("Dns hosts domain stays unchanged as a whole")
		return nil
	}
	// Create new domains records
	if len(PodDomainInfoToHosts) > 0 {
		for domain, ip := range PodDomainInfoToHosts {
			klog.V(4).Infof("Adding dns hosts domain %s and ip %s", domain, ip)
			h.hostsMap[domain] = ip
		}
	}
	// Sync dns hosts since it has changed now
	if err := h.saveHosts(); err != nil {
		return err
	}
	return nil
}

func (h *Hosts) SetHostsByMap(hostsMap map[string]string) error {
	h.Lock()
	defer h.Unlock()
	if !reflect.DeepEqual(h.hostsMap, hostsMap) {
		originalHostsMap := h.hostsMap
		h.hostsMap = hostsMap
		if err := h.saveHosts(); err != nil {
			h.hostsMap = originalHostsMap
			klog.V(4).Infof("Reset dns hosts domain and ip as a whole err %v", err)
			return err
		}
		klog.V(4).Infof("Reset dns hosts domain and ip as a whole successfully")
	}
	return nil
}

func (h *Hosts) saveHosts() error {
	hostData := []byte(h.parseHostsToFile())
	err := ioutil.WriteFile(h.hostPath, hostData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (h *Hosts) parseHostsToFile() string {
	hf := ""
	for domain, ip := range h.hostsMap {
		hf = hf + fmt.Sprintln(fmt.Sprintf("%s %s", ip, domain))
	}
	return hf
}

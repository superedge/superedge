package util

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
)

var suffix = []string{
	".svc.cluster.local",
	".svc.cluster",
	".svc",
}

type Hosts struct {
	HostPath string
	HostsMap map[string][]string
	sync.Mutex
}

func NewHosts(HostPath string) Hosts {
	return Hosts{
		HostPath: HostPath,
		HostsMap: make(map[string][]string),
	}
}

func (h *Hosts) ReadHostsFile(HostsPath string) ([]byte, error) {
	bs, err := ioutil.ReadFile(HostsPath)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (h *Hosts) ParseHosts(hostsFileContent []byte, err error) (map[string][]string, error) {
	if err != nil {
		return nil, err
	}
	h.Lock()
	defer h.Unlock()
	for _, line := range strings.Split(strings.Trim(string(hostsFileContent), " \t\r\n"), "\n") {
		line = strings.Replace(strings.Trim(line, " \t"), "\t", " ", -1)
		if len(line) == 0 || line[0] == ';' || line[0] == '#' {
			continue
		}
		pieces := strings.SplitN(line, " ", 2)
		//if len(pieces) > 1 && len(pieces[0]) > 0 {
		if names := strings.Fields(pieces[1]); len(names) > 0 {
			if _, ok := h.HostsMap[pieces[0]]; ok {
				h.HostsMap[pieces[0]] = append(h.HostsMap[pieces[0]], names...)
			} else {
				h.HostsMap[pieces[0]] = names
			}
		}
	}
	return h.HostsMap, nil
}

func (h *Hosts) UpdateHosts(PodInfoToHost map[string]string, nameSpace, setName, svcName string) error {
	h.Lock()
	defer h.Unlock()

	for ip, hosts := range h.HostsMap {
		temp_hosts := []string{}
		for _, host := range hosts {
			match, _ := regexp.MatchString(setName+"-"+`[0-9]+`+`\.`+nameSpace+`\.`+svcName+`.*`, host)
			if !match {
				//klog.Infof("%s not match", host)
				temp_hosts = append(temp_hosts, host)
			}
		}
		if len(temp_hosts) == 0 {
			delete(h.HostsMap, ip)
		} else {
			h.HostsMap[ip] = temp_hosts
		}
	}

	for addIp, addHost := range PodInfoToHost {
		hostWithNsSvc := addHost + "." + nameSpace + "." + svcName
		hostList := []string{hostWithNsSvc}
		for _, suf := range suffix {
			hostList = append(hostList, hostWithNsSvc+suf)
		}

		if _, ok := h.HostsMap[addIp]; ok {
			h.HostsMap[addIp] = append(h.HostsMap[addIp], hostList...)
		} else {
			h.HostsMap[addIp] = hostList
		}
	}

	if err := h.Save(); err != nil {
		return err
	}
	return nil
}

func (h *Hosts) Save() error {
	hostData := []byte(h.ParseHostsFile())

	h.Lock()
	defer h.Unlock()

	err := ioutil.WriteFile(h.HostPath, hostData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (h *Hosts) ParseHostsFile() string {
	h.Lock()
	defer h.Unlock()

	hf := ""

	for ip, host := range h.HostsMap {
		hf = hf + fmt.Sprintln(fmt.Sprintf("%s %s", ip, strings.Join(host, " ")))
	}

	return hf
}

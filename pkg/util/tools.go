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

package util

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func GetHostAllIPs() (map[string]string, error) {
	ips := make(map[string]string)

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		byName, err := net.InterfaceByName(i.Name)
		if err != nil {
			return nil, err
		}
		addresses, err := byName.Addrs()
		for _, v := range addresses {
			ips[byName.Name] = v.String()
		}
	}
	return ips, nil
}

func GetLocalIP() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addresses {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() &&
			!ipNet.IP.IsLinkLocalUnicast() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", errors.New("Not found Local IP\n")
}

func GetHostPublicIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if localAddr.String() != "" {
		ips := strings.Split(localAddr.String(), ":")
		if 2 == len(ips) {
			return ips[0], nil
		}
	}
	return "", errors.New("Not found Local IP\n")
}

func GetIPByInterfaceName(name string) (ip string, err error) {
	byName, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addresses, err := byName.Addrs()
	if err != nil {
		return "", err
	}

	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				return ip, nil
			}
		}
	}
	return ip, nil
}

func GetLocalAddrByInterface(name string) (net.Addr, error) {
	localIP, err := GetIPByInterfaceName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get local ip, err: %v", err)
	}
	// localIP，":0" indicates that the port is automatically selected
	localAddr := &net.TCPAddr{
		IP:   net.ParseIP(localIP),
		Port: 0,
	}
	return localAddr, nil
}

func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str, end)
	return str[s:e]
}

func SplitHostPortIgnoreMissingPort(v string) (addr string, port string, err error) {
	addr, port, err = net.SplitHostPort(v)
	if err != nil {
		if aerr, ok := err.(*net.AddrError); ok {
			if strings.HasPrefix(aerr.Err, "missing port") {
				// ignore missing port number.
				addr, port, err = v, "", nil
			}
		}
	}
	return addr, port, err
}

func SplitHostPortWithDefaultPort(v string, defaultPort string) (addr string, port string, err error) {
	addr, port, err = net.SplitHostPort(v)
	if err != nil {
		if aerr, ok := err.(*net.AddrError); ok {
			if strings.HasPrefix(aerr.Err, "missing port") {
				// ignore missing port number.
				addr, port, err = net.SplitHostPort(v + ":" + defaultPort)
			}
		}
	}
	return addr, port, err
}

func RemoveDuplicateElement(slice []string) []string {
	result := make([]string, 0, len(slice))
	temp := map[string]struct{}{}
	for _, item := range slice {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func DeleteSliceElement(slice []string, element string) []string {
	var ret []string
	for _, val := range slice {
		if val != element {
			ret = append(ret, val)
		}
	}
	return ret
}

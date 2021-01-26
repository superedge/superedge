package util

import (
	"errors"
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

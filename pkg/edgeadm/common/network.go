package common

import (
	"fmt"
	"net"

	"github.com/superedge/superedge/pkg/util/ipallocator"
)

func GetIndexedIP(subnet string, index int) (net.IP, error) {
	_, svcSubnetCIDR, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse service subnet CIDR %q, error %s", subnet, err)
	}

	dnsIP, err := ipallocator.GetIndexedIP(svcSubnetCIDR, index)
	if err != nil {
		return nil, fmt.Errorf("unable to get %dth IP address from service subnet CIDR %s, error: %s", index, svcSubnetCIDR.String(), err)
	}

	return dnsIP, nil
}

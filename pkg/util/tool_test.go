package util

import (
	"fmt"
	"testing"
)

func TestIps(t *testing.T) {
	fmt.Println(GetHostAllIPs())
}

func TestGetLocalIP(t *testing.T) {
	fmt.Println(GetLocalIP())
}

func TestGetHostUSingIP(t *testing.T) {
	fmt.Println(GetHostPublicIP())
}

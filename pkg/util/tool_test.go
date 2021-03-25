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

func TestWriteWithAppend(t *testing.T) {
	name := "./test.txt"
	content := "12345\n"
	fmt.Println(WriteWithAppend(name, content))
}

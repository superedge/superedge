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

func TestRunLinuxCommand(t *testing.T) {
	stdout, stderr, err := RunLinuxCommand("ls -l")
	fmt.Println("stdout:", stdout, "stderr:", stderr, "err:", err)
}

func TestRunLinuxShellFile(t *testing.T) {
	stdout, stderr, err := RunLinuxShellFile("./shell_test.sh")
	fmt.Println("stdout:", stdout, "stderr:", stderr, "err:", err)
}

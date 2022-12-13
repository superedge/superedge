package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
)

var (
	action  string
	service string
)

func init() {
	if len(os.Args) >= 3 && len(os.Args) <= 2 {
		fmt.Printf("usage: %s <action> <service-name>\n action can be [start, stop, restart, status]\n or %s daemon-reload", os.Args[0], os.Args[0])
		os.Exit(1)
	}
	if len(os.Args) == 2 {
		action = os.Args[1]
	}
	if len(os.Args) == 3 {
		action = os.Args[1]
		service = os.Args[2]
		if !strings.HasSuffix(service, ".service") {
			service = fmt.Sprintf("%s.service", service)
		}
	}
}
func main() {
	systemctl, err := dbus.New()
	if err != nil {
		fmt.Printf("Create systemctl failed: %s\n", err.Error())
		os.Exit(1)
	}
	if action == "status" {
		units, err := systemctl.ListUnits()
		if err != nil {
			fmt.Printf("systemctl status %s failed: %s\n", service, err.Error())
			os.Exit(1)
		}
		for _, unit := range units {
			if unit.Name == service {
				if unit.ActiveState == "active" {
					fmt.Printf("service status is active\n")
					return
				} else {
					fmt.Printf("service status is %s\n", unit.ActiveState)
					os.Exit(1)
				}
			}
		}
		fmt.Printf("no such service: %s\n", service)
		os.Exit(1)
	}
	if action == "daemon-reload" {
		if err := systemctl.Reload(); err != nil {
			fmt.Printf("daemon-reload err %v", err)
			os.Exit(1)
		}
		return
	}
	f := func(action string) func(string, string, chan<- string) (int, error) {
		switch action {
		case "start":
			return systemctl.StartUnit
		case "stop":
			return systemctl.StopUnit
		case "restart":
			return systemctl.RestartUnit
		default:
			fmt.Printf("no such action: %s\n", action)
			os.Exit(1)
		}
		return nil
	}(action)
	ch := make(chan string)
	defer close(ch)
	if _, err := f(service, "replace", ch); err != nil {
		fmt.Printf("systemctl %s %s failed: %s\n", action, service, err.Error())
		os.Exit(1)
	}
	ans := <-ch
	if ans != "done" {
		fmt.Printf("systemctl %s %s failed: %s\n", action, service, ans)
		os.Exit(1)
	}
	fmt.Printf("systemctl %s %s successed\n", action, service)
}

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
	"bufio"
	"bytes"
	"github.com/tatsushid/go-fastping"
	"io"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	RequestCache = 10 * 1024
)

func ReplaceString(line string) string {
	line = strings.Replace(line, " ", "", -1)
	line = strings.Replace(line, "\n", "", -1)
	return line
}

func GetRequestFromConn(conn net.Conn) (*http.Request, *bytes.Buffer, error) {
	rawRequest := bytes.NewBuffer(make([]byte, RequestCache))
	rawRequest.Reset()
	reqReader := bufio.NewReader(io.TeeReader(conn, rawRequest))
	request, err := http.ReadRequest(reqReader)
	if err != nil {
		klog.Errorf("Failed to get http request, error: %v", err)
		return nil, nil, err
	}
	return request, rawRequest, nil
}

func Ping(ip string) error {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		klog.Errorf("Failed to get icmp address, ip %s, error: %v", ip, err)
		return err
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
	}
	p.OnIdle = func() {
	}
	return p.Run()
}

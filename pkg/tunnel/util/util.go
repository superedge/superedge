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
		klog.V(8).Infof("Failed to get http request, error: %v", err)
		return nil, nil, err
	}
	return request, rawRequest, nil
}

func GetRespFromConn(conn net.Conn, req *http.Request) (*http.Response, *bytes.Buffer, error) {
	rawResp := bytes.NewBuffer(make([]byte, RequestCache))
	rawResp.Reset()
	reqReader := bufio.NewReader(io.TeeReader(conn, rawResp))
	resp, err := http.ReadResponse(reqReader, req)
	if err != nil {
		klog.V(8).Infof("Failed to get http request, error: %v", err)
		return nil, nil, err
	}
	return resp, rawResp, nil
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

func WriteResponseMsg(conn net.Conn, respMsg, tranceId, status string, statusCode int) error {
	resp := http.Response{
		Status:     status,
		StatusCode: statusCode,
		Header: map[string][]string{
			STREAM_TRACE_ID: {tranceId},
		},
		Body:          io.NopCloser(strings.NewReader(respMsg)),
		ContentLength: int64(len(respMsg)),
	}
	err := resp.Write(conn)
	if err != nil {
		klog.ErrorS(err, "failed to write response data  to proxyConn", "response data", respMsg, STREAM_TRACE_ID, tranceId)
		return err
	}
	return nil
}

func InternalServerErrorMsg(proxyConn net.Conn, respMsg, tranceId string) error {
	return WriteResponseMsg(proxyConn, respMsg, tranceId, "Internal Server Error", http.StatusInternalServerError)
}

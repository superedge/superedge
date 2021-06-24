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

package httpsmng

import (
	"crypto/tls"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/context"
	"github.com/superedge/superedge/pkg/tunnel/proto"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

type ServerHandler struct {
	port string
}

func StartServer() {
	if conf.TunnelConf.TunnlMode.Cloud.Https == nil {
		return
	}
	cert, err := tls.LoadX509KeyPair(conf.TunnelConf.TunnlMode.Cloud.Https.Cert, conf.TunnelConf.TunnlMode.Cloud.Https.Key)
	if err != nil {
		klog.Errorf("client load cert fail certpath = %s keypath = %s \n", conf.TunnelConf.TunnlMode.Cloud.Https.Cert, conf.TunnelConf.TunnlMode.Cloud.Https.Key)
		return
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	for k := range conf.TunnelConf.TunnlMode.Cloud.Https.Addr {
		serverHandler := &ServerHandler{
			port: k,
		}
		s := &http.Server{
			Addr:      "0.0.0.0:" + k,
			Handler:   serverHandler,
			TLSConfig: config,
		}
		klog.Infof("the https server of the cloud tunnel listen on %s", s.Addr)
		go func(server *http.Server) {
			err = s.ListenAndServeTLS("", "")
			if err != nil {
				klog.Errorf("server start fail,add = %s err = %v", s.Addr, err)
			}
		}(s)
	}
}
func (serverHandler *ServerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var nodeName string
	nodeinfo := strings.Split(request.Host, ":")
	if context.GetContext().NodeIsExist(nodeinfo[0]) {
		nodeName = nodeinfo[0]
	} else {
		nodeName = request.TLS.ServerName
	}
	node := context.GetContext().GetNode(nodeName)
	if node == nil {
		fmt.Fprintf(writer, "edge node disconnected node = %s", nodeinfo[0])
		return
	}
	uid := uuid.NewV4().String()
	node.BindNode(uid)
	conn := context.GetContext().AddConn(uid)

	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		klog.Errorf("traceid = %s read request body fail err = %v ", uid, err)
		fmt.Fprintf(writer, "traceid = %s read request body fail err = %v ", uid, err)
		return
	}
	httpmsg := &HttpsMsg{
		HttpsStatus: util.CONNECTING,
		Header:      make(map[string][]string),
		Method:      request.Method,
		HttpBody:    requestBody,
	}

	httpmsg.Header = request.Header

	bmsg := httpmsg.Serialization()
	if len(bmsg) == 0 {
		klog.Errorf("traceid = %s httpsmsg serialization failed err = %v req = %v serverName = %s", uid, err, request, request.TLS.ServerName)
		fmt.Fprintf(writer, "traceid = %s httpsmsg serialization failed err = %v", uid, err)
		return
	}
	node.Send2Node(&proto.StreamMsg{
		Node:     nodeName,
		Category: util.HTTPS,
		Type:     util.CONNECTING,
		Topic:    uid,
		Data:     bmsg,
		Addr:     "https://" + conf.TunnelConf.TunnlMode.Cloud.Https.Addr[serverHandler.port] + request.URL.String(),
	})
	if err != nil {
		klog.Errorf("traceid = %s httpsServer send request msg failed err = %v", uid, err)
		fmt.Fprintf(writer, "traceid = %s httpsServer send request msg failed err = %v", uid, err)
		return
	}
	resp := <-conn.ConnRecv()
	rmsg, err := Deserialization(resp.Data)
	if err != nil {
		klog.Errorf("traceid = %s httpsmag deserialization failed err = %v", uid, err)
		fmt.Fprintf(writer, "traceid = %s httpsmag deserialization failed err = %v", uid, err)
		return
	}
	node.Send2Node(&proto.StreamMsg{
		Node:     nodeName,
		Category: util.HTTPS,
		Type:     util.CONNECTED,
		Topic:    uid,
	})
	if err != nil {
		klog.Errorf("traceid = %s httpsServer send confirm msg failed err = %v", uid, err)
		fmt.Fprintf(writer, "traceid = %s httpsServer send confirm msg failed err = %v", uid, err)
		return
	}
	if rmsg.StatusCode != http.StatusSwitchingProtocols {
		handleServerHttp(rmsg, writer, request, node, conn)
	} else {
		handleServerSwitchingProtocols(writer, node, conn)
	}
}

func handleServerHttp(rmsg *HttpsMsg, writer http.ResponseWriter, request *http.Request, node context.Node, conn context.Conn) {
	for k, v := range rmsg.Header {
		for _, vv := range v {
			writer.Header().Add(k, vv)
		}
	}
	flusher, ok := writer.(http.Flusher)
	if ok {
		running := true
		for running {
			select {
			case <-request.Context().Done():
				klog.Infof("traceid = %s httpServer context close! ", conn.GetUid())
				node.Send2Node(&proto.StreamMsg{
					Node:     node.GetName(),
					Category: util.HTTPS,
					Type:     util.CLOSED,
					Topic:    conn.GetUid(),
				})
				running = false
			case msg := <-conn.ConnRecv():
				if msg.Data != nil && len(msg.Data) != 0 {
					_, err := writer.Write(msg.Data)
					if err != nil {
						klog.Errorf("traceid = %s httpsServer write data failed err = %v", conn.GetUid(), err)
					}
					flusher.Flush()
				}
				if msg.Type == util.CLOSED {
					running = false
					break
				}
			}
		}
	}
	context.GetContext().RemoveConn(conn.GetUid())
}

func handleServerSwitchingProtocols(writer http.ResponseWriter, node context.Node, conn context.Conn) {
	requestHijacker, ok := writer.(http.Hijacker)
	if !ok {
		klog.Errorf("traceid = %s unable to hijack response writer: %T", conn.GetUid(), writer)
		fmt.Fprintf(writer, "traceid = %s unable to hijack response writer: %T", conn.GetUid(), writer)
		return
	}
	requestHijackedConn, _, err := requestHijacker.Hijack()
	if err != nil {
		klog.Errorf("traceid = %s unable to hijack response: %v", conn.GetUid(), err)
		fmt.Fprintf(writer, "traceid = %s unable to hijack response: %v", conn.GetUid(), err)
		return
	}
	writerComplete := make(chan struct{})
	readerComplete := make(chan struct{})
	stop := make(chan struct{}, 1)
	go NetRead(requestHijackedConn, conn.GetUid(), node, stop, readerComplete)
	go NetWrite(requestHijackedConn, node, conn, stop, writerComplete)
	select {
	case <-writerComplete:
	case <-readerComplete:
	}
	klog.Infof("traceid = %s httpsServer close ", conn.GetUid())
}

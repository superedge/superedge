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
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"k8s.io/klog"
	"net"
	"net/http"
	"superedge/pkg/tunnel/conf"
	"superedge/pkg/tunnel/context"
	"superedge/pkg/tunnel/proto"
	"superedge/pkg/tunnel/util"
)

func Request(msg *proto.StreamMsg) {
	httpConn, err := getHttpConn(msg)
	if err != nil {
		klog.Errorf("traceid = %s failed to get httpclient httpConn err = %v", msg.Topic, err)
		return
	}
	rawResponse := bytes.NewBuffer(make([]byte, 0, util.MaxResponseSize))
	rawResponse.Reset()
	respReader := bufio.NewReader(io.TeeReader(httpConn, rawResponse))
	resp, err := http.ReadResponse(respReader, nil)
	if err != nil {
		klog.Errorf("traceid = %s httpsclient read response failed err = %v", msg.Topic, err)
		return
	}

	bodyMsg := HttpsMsg{
		StatusCode:  resp.StatusCode,
		HttpsStatus: util.CONNECTED,
		Header:      make(map[string]string),
	}
	for k, v := range resp.Header {
		for _, vv := range v {
			bodyMsg.Header[k] = vv
		}
	}
	msgData := bodyMsg.Serialization()
	if len(msgData) == 0 {
		klog.Errorf("traceid = %s httpsclient httpsmsg serialization failed", msg.Topic)
		return
	}
	node := context.GetContext().GetNode(msg.Node)
	if node == nil {
		klog.Errorf("traceid = %s httpClient failed to get node", msg.Topic)
		return
	}
	node.Send2Node(&proto.StreamMsg{
		Node:     msg.Node,
		Category: msg.Category,
		Type:     util.CONNECTED,
		Topic:    msg.Topic,
		Data:     msgData,
	})
	conn := context.GetContext().AddConn(msg.Topic)
	node.BindNode(msg.Topic)
	confirm := true
	for confirm {
		confirmMsg := <-conn.ConnRecv()
		if confirmMsg.Type == util.CONNECTED {
			confirm = false
		}
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		handleClientHttp(resp, rawResponse, httpConn, msg, node, conn)
	} else {
		handleClientSwitchingProtocols(httpConn, rawResponse, msg, node, conn)
	}
}

func getHttpConn(msg *proto.StreamMsg) (net.Conn, error) {
	cert, err := tls.LoadX509KeyPair(conf.TunnelConf.TunnlMode.EDGE.Https.Cert, conf.TunnelConf.TunnlMode.EDGE.Https.Key)
	if err != nil {
		klog.Errorf("tranceid = %s httpsclient load cert fail certpath = %s keypath = %s", msg.Topic, conf.TunnelConf.TunnlMode.EDGE.Https.Cert, conf.TunnelConf.TunnlMode.EDGE.Https.Key)
		return nil, err
	}
	requestMsg, err := Deserialization(msg.Data)
	if err != nil {
		klog.Errorf("traceid = %s httpsclient deserialization failed err = %v", msg.Topic, err)
		return nil, err
	}
	request, err := http.NewRequest(requestMsg.Method, msg.Addr, nil)
	if err != nil {
		klog.Errorf("traceid = %s httpsclient get request fail err = %v", msg.Topic, err)
		return nil, err
	}
	for k, v := range requestMsg.Header {
		request.Header.Add(k, v)
	}
	conn, err := tls.Dial("tcp", request.Host, &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	})
	if err != nil {
		klog.Errorf("traceid = %s httpsclient request failed err = %v", msg.Topic, err)
		return nil, err
	}
	err = request.Write(conn)
	if err != nil {
		klog.Errorf("traceid = %s https clinet request failed to write conn err = %v", msg.Topic, err)
		return nil, err
	}
	return conn, nil
}

func handleClientHttp(resp *http.Response, rawResponse *bytes.Buffer, httpConn net.Conn, msg *proto.StreamMsg, node context.Node, conn context.Conn) {
	readCh := make(chan *proto.StreamMsg, util.MSG_CHANNEL_CAP)
	stop := make(chan struct{})
	go func(read chan *proto.StreamMsg, response *http.Response, buf *bytes.Buffer, stopRead chan struct{}) {
		rrunning := true
		for rrunning {
			bbody := make([]byte, util.MaxResponseSize)
			n, err := response.Body.Read(bbody)
			respMsg := &proto.StreamMsg{
				Node:     msg.Node,
				Category: msg.Category,
				Type:     util.CONNECTED,
				Topic:    msg.Topic,
				Data:     bbody[:n],
			}
			if err != nil {
				if err == io.EOF {
					klog.V(4).Infof("traceid = %s httpsclient read fail err = %v", msg.Topic, err)
				} else {
					klog.Errorf("traceid = %s httpsclient read fail err = %v", msg.Topic, err)
				}
				rrunning = false
				respMsg.Type = util.CLOSED
			} else {
				respMsg.Type = util.TRANSNMISSION
				buf.Reset()
			}
			read <- respMsg
		}
		<-stop
		close(read)
	}(readCh, resp, rawResponse, stop)
	running := true
	for running {
		select {
		case cloudMsg := <-conn.ConnRecv():
			if cloudMsg.Type == util.CLOSED {
				klog.Infof("traceid = %s httpsclient receive close msg", msg.Topic)
				httpConn.Close()
				stop <- struct{}{}
			}
		case respMsg := <-readCh:
			if respMsg == nil {
				running = false
				break
			}
			node.Send2Node(respMsg)
			if respMsg.Type == util.CLOSED {
				stop <- struct{}{}
				klog.V(4).Infof("traceid = %s httpsclient read fail !", msg.Topic)
				running = false
			}

		}
	}
	node.UnbindNode(conn.GetUid())
	context.GetContext().RemoveConn(conn.GetUid())
}

func handleClientSwitchingProtocols(httpConn net.Conn, rawResponse *bytes.Buffer, msg *proto.StreamMsg, node context.Node, conn context.Conn) {
	node.Send2Node(&proto.StreamMsg{
		Node:     msg.Node,
		Category: util.HTTPS,
		Type:     util.TRANSNMISSION,
		Topic:    msg.Topic,
		Data:     rawResponse.Bytes(),
	})
	writerComplete := make(chan struct{})
	readerComplete := make(chan struct{})
	stop := make(chan struct{}, 1)
	go IoRead(httpConn, msg.Topic, node, stop, readerComplete)
	go IoWrite(httpConn, node, conn, stop, writerComplete)
	select {
	case <-writerComplete:
	case <-readerComplete:
	}
}

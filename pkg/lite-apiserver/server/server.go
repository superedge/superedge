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

package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/superedge/superedge/pkg/util"
	"io/ioutil"
	"net/http"

	"github.com/superedge/superedge/cmd/lite-apiserver/app/options"
	"github.com/superedge/superedge/pkg/lite-apiserver/cert"
	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/proxy"
	"github.com/superedge/superedge/pkg/lite-apiserver/transport"

	"k8s.io/klog"
)

type LiteServer struct {
	ServerConfig *config.LiteServerConfig
	stopCh       <-chan struct{}
}

func CreateServer(serverOptions *options.ServerRunOptions, stopCh <-chan struct{}) (*LiteServer, error) {

	config, err := serverOptions.ApplyTo()
	if err != nil {
		return nil, err
	}
	return &LiteServer{
		ServerConfig: config,
		stopCh:       stopCh,
	}, nil
}

func (s *LiteServer) Run() error {

	// init cert manager
	certManager := cert.NewCertManager(s.ServerConfig)
	err := certManager.Init()
	if err != nil {
		klog.Errorf("Init certManager error: %v", err)
		return err
	}
	certManager.Start()

	// init transport manager
	transportManager := transport.NewTransportManager(s.ServerConfig, certManager)
	err = transportManager.Init()
	if err != nil {
		klog.Errorf("Init transportManager error: %v", err)
		return err
	}

	cacher := proxy.NewRequestCacheController(s.ServerConfig, transportManager)
	go cacher.Run(s.stopCh)

	edgeServerHandler, err := proxy.NewEdgeServerHandler(s.ServerConfig, transportManager, cacher)
	if err != nil {
		klog.Errorf("Create edgeServerHandler error: %v", err)
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", edgeServerHandler)
	mux.HandleFunc("/debug/flags/v", util.UpdateLogLevel)

	caCrt, err := ioutil.ReadFile(s.ServerConfig.CAFile)
	if err != nil {
		klog.Errorf("Read ca file err: %v", err)
		return err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCrt)

	ser := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.ServerConfig.Port),
		Handler: mux,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.VerifyClientCertIfGiven,
		},
	}
	go func() {
		klog.Infof("Listen on %s", ser.Addr)
		klog.Fatal(ser.ListenAndServeTLS(s.ServerConfig.CertFile, s.ServerConfig.KeyFile))
	}()

	<-s.stopCh
	klog.Info("Received a program exit signal")
	return nil
}

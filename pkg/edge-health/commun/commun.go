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

package commun

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health/config"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"github.com/superedge/superedge/pkg/edge-health/util"
	pkgutil "github.com/superedge/superedge/pkg/util"
	"io"
	"k8s.io/apimachinery/pkg/util/wait"
	corelisters "k8s.io/client-go/listers/core/v1"
	log "k8s.io/klog"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Commun interface {
	Commun(*metadata.CheckMetadata, corelisters.ConfigMapLister, string, <-chan struct{})
}

type CommunEdge struct {
	*config.EdgeHealthCommun
	client *http.Client
}

func NewCommunEdge(cfg *config.EdgeHealthCommun) *CommunEdge {
	return &CommunEdge{
		EdgeHealthCommun: cfg,
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: time.Duration(cfg.CommunTimeout) * time.Second,
		},
	}
}

func (c *CommunEdge) Commun(checkMetadata *metadata.CheckMetadata, cmLister corelisters.ConfigMapLister, localIp string, stopCh <-chan struct{}) {
	go c.communReceive(checkMetadata, cmLister, stopCh)
	wait.Until(func() {
		c.communSend(checkMetadata, cmLister, localIp)
	}, time.Duration(c.CommunPeriod)*time.Second, stopCh)
}

// TODO: support changeable server listen port
func (c *CommunEdge) communReceive(checkMetadata *metadata.CheckMetadata, cmLister corelisters.ConfigMapLister, stopCh <-chan struct{}) {
	svr := &http.Server{Addr: ":" + strconv.Itoa(c.CommunServerPort)}
	svr.ReadTimeout = time.Duration(c.CommunTimeout) * time.Second
	svr.WriteTimeout = time.Duration(c.CommunTimeout) * time.Second
	http.HandleFunc("/debug/flags/v", pkgutil.UpdateLogLevel)
	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		var communInfo metadata.CommunInfo
		if r.Body == nil {
			http.Error(w, "Invalid commun information", http.StatusBadRequest)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&communInfo)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid commun information %+v", err), http.StatusBadRequest)
			return
		}
		log.V(4).Infof("Received common information from %s : %+v", communInfo.SourceIP, communInfo.CheckDetail)

		if _, err := io.WriteString(w, "Received!\n"); err != nil {
			log.Errorf("communReceive: send response err %+v", err)
			http.Error(w, fmt.Sprintf("Send response err %+v", err), http.StatusInternalServerError)
			return
		}
		if hmac, err := util.GenerateHmac(communInfo, cmLister); err != nil {
			log.Errorf("communReceive: server GenerateHmac err %+v", err)
			http.Error(w, fmt.Sprintf("GenerateHmac err %+v", err), http.StatusInternalServerError)
			return
		} else {
			if hmac != communInfo.Hmac {
				log.Errorf("communReceive: Hmac not equal, hmac is %s but received commun info hmac is %s", hmac, communInfo.Hmac)
				http.Error(w, "Hmac not match", http.StatusForbidden)
				return
			}
		}
		log.V(4).Infof("communReceive: Hmac match")

		checkMetadata.SetByCommunInfo(communInfo)
		log.V(4).Infof("After communicate, check info is %+v", checkMetadata.CheckInfo)
	})

	go func() {
		if err := svr.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server: exit with error %+v", err)
		}
	}()

	for {
		select {
		case <-stopCh:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := svr.Shutdown(ctx); err != nil {
				log.Errorf("Server: program exit, server exit error %+v", err)
			}
			return
		default:
		}
	}
}

func (c *CommunEdge) communSend(checkMetadata *metadata.CheckMetadata, cmLister corelisters.ConfigMapLister, localIp string) {
	copyLocalCheckDetail := checkMetadata.CopyLocal(localIp)
	var checkedIps []string
	for checkedIp := range copyLocalCheckDetail {
		checkedIps = append(checkedIps, checkedIp)
	}
	util.ParallelizeUntil(context.TODO(), 16, len(checkedIps), func(index int) {
		// Only send commun information to other edge nodes(excluding itself)
		dstIp := checkedIps[index]
		if dstIp == localIp {
			return
		}
		// Send commun information
		communInfo := metadata.CommunInfo{SourceIP: localIp, CheckDetail: copyLocalCheckDetail}
		if hmac, err := util.GenerateHmac(communInfo, cmLister); err != nil {
			log.Errorf("communSend: generateHmac err %+v", err)
			return
		} else {
			communInfo.Hmac = hmac
		}
		commonInfoBytes, err := json.Marshal(communInfo)
		if err != nil {
			log.Errorf("communSend: json.Marshal commun info err %+v", err)
			return
		}
		commonInfoReader := bytes.NewReader(commonInfoBytes)
		for i := 0; i < c.CommunRetries; i++ {
			req, err := http.NewRequest("PUT", "http://"+dstIp+":"+strconv.Itoa(c.CommunServerPort)+"/result", commonInfoReader)
			if err != nil {
				log.Errorf("communSend: NewRequest for remote edge node %s err %+v", dstIp, err)
				continue
			}
			if err = util.DoRequestAndDiscard(c.client, req); err != nil {
				log.Errorf("communSend: DoRequestAndDiscard for remote edge node %s err %+v", dstIp, err)
			} else {
				log.V(4).Infof("communSend: put commun info %+v to remote edge node %s successfully", communInfo, dstIp)
				break
			}
		}
	})
}

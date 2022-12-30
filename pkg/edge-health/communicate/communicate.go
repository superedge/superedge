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

package communicate

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/superedge/superedge/pkg/edge-health/check"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/data"
	"github.com/superedge/superedge/pkg/edge-health/util"
	pkgutil "github.com/superedge/superedge/pkg/util"
	"k8s.io/klog/v2"
)

type Communicate interface {
	Server(ctx context.Context, wg *sync.WaitGroup)
	Client()
	GetPeriod() int
}

func NewCommunicateEdge(communicatePeriod, communicateTimetout, communicateRetryTime, communicateServerPort int) Communicate {
	pm := NewProberManager(int32(communicateTimetout))

	return &CommunicateEdge{
		communicatePeriod,
		communicateTimetout,
		communicateRetryTime,
		communicateServerPort,
		pm,
		buildSourceInfo(),
	}
}

type CommunicateEdge struct {
	CommunicatePeriod     int
	CommunicateTimeout    int
	CommunicateRetryTime  int
	CommunicateServerPort int
	Prober                *ProberManager
	SourceInfo            *SourceInfo
}

func (c *CommunicateEdge) GetPeriod() int {
	return c.CommunicatePeriod
}

func (c *CommunicateEdge) HandleProbe(w http.ResponseWriter, r *http.Request) {

	var probe Probe
	if r.Body == nil {
		http.Error(w, "Invalid Body", http.StatusPaymentRequired)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.ErrorS(err, "Read request body error", "request url", r.URL.String())
		http.Error(w, "Invalid Body", http.StatusPaymentRequired)
		return
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		klog.ErrorS(err, "Decode request body error", "request url", string(body))
		http.Error(w, "Invalid Body", http.StatusPaymentRequired)
		return
	}
	klog.V(6).InfoS("probe request info", "req", string(body))
	// validate target
	for _, t := range probe.Targets {
		if t.Protocol != "" {
			if _, ok := ProtoSet[t.Protocol]; !ok {
				klog.ErrorS(err, "Invalid Protocol", "target ip", t.IP, "target port", t.Port, "target protocol", t.Protocol)
				http.Error(w, "Invalid Protocol "+t.Protocol, http.StatusBadRequest)
				return
			}
		}
		if len(t.Name) > MaxTargetName {
			klog.ErrorS(err, "Invalid Target Name", "target name", t.Name)
			http.Error(w, "Invalid Target Name"+t.Name, http.StatusBadRequest)
			return
		}
	}

	pieces := len(probe.Targets)
	// if timeout is 1s, 1 worker will address 5 piece, that mean
	// the whole request max response is 5s
	workers := (pieces / 5) + 1

	if pieces > MaxProbeIPs {
		http.Error(w, "Targets Too Long", http.StatusBadRequest)
		return
	}
	c.Prober.Parallelize(workers, pieces, func(i int) {
		klog.V(6).InfoS("get piece index", "index", i)
		t := probe.Targets[i]
		t.Normal = new(bool)
		switch t.Protocol {
		case ProtoTCP:
			*(t.Normal) = c.Prober.TCPProbe(t.IP, t.Port, c.Prober.ProbeTimeout)
		case ProtoICMP:
			*(t.Normal) = c.Prober.ICMPProbe(t.IP, c.Prober.ProbeTimeout)
		}
	})
	probeResp := &ProbeResp{
		Targets:    probe.Targets,
		SourceInfo: c.SourceInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(probeResp)
	if err != nil {
		klog.ErrorS(err, "Marshal response body error", "request url", r.URL.String())
		http.Error(w, "Invalid Body", http.StatusInternalServerError)
	}
	klog.V(6).InfoS("probe response info", "resp", string(resp))
	w.Write(resp)
}

func buildSourceInfo() *SourceInfo {
	return &SourceInfo{
		SourcePodIP:    common.PodIP,
		SourcePodName:  common.PodName,
		SourceNodeIP:   common.NodeIP,
		SourceNodeName: common.NodeName,
	}
}

//TODO: 监听端口可变
func (c *CommunicateEdge) Server(ctx context.Context, wg *sync.WaitGroup) {
	mux := http.NewServeMux()
	mux.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		var communicatedata data.CommunicateData
		if r.Body == nil {
			http.Error(w, "Invalid Body", http.StatusPaymentRequired)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&communicatedata)
		if err != nil {
			klog.ErrorS(err, "Decode request body error", "request url", r.URL.String())
			http.Error(w, "Invalid Body", http.StatusPaymentRequired)
			return
		}

		klog.V(4).Infof("Communicate Server: received data from %v : %v", communicatedata.SourceIP, communicatedata.ResultDetail)
		if hmac, err := util.GenerateHmac(communicatedata); err != nil {
			klog.Errorf("Communicate Server: server GenerateHmac err: %v", err)
			http.Error(w, "Hmac not match", http.StatusInternalServerError)
			return
		} else {
			if hmac != communicatedata.Hmac {
				klog.Errorf("Communicate Server: Hmac not equal, hmac is %s, communicatedata.Hmac is %s", hmac, communicatedata.Hmac)
				http.Error(w, "Hmac not match", http.StatusForbidden)
				return
			}
		}
		klog.V(4).Infof("Communicate Server: Hmac match")

		data.Result.SetResult(&communicatedata)
		klog.V(6).Infof("After communicate, result is %v", data.Result.GetResultDataAll())
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/debug/flags/v", pkgutil.UpdateLogLevel)
	mux.HandleFunc("/probe", c.HandleProbe)
	mux.HandleFunc("/localinfo", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("fullmesh") != "" {
			localInfoData := data.Result.GetResultDataAll()
			if err := json.NewEncoder(w).Encode(localInfoData); err != nil {
				klog.Errorf("Get local info all:  err: %v", err)
				return
			}
		}
		localInfoData := data.Result.CopyLocalResultData(common.NodeIP)
		if err := json.NewEncoder(w).Encode(localInfoData); err != nil {
			klog.Errorf("Get Local Info: NodeInfo err: %v", err)
			return
		}
	})

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(c.CommunicateServerPort),
		Handler: http.Handler(mux),
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			klog.Fatalf("Server: exit with error: %v", err)
		}
	}()

	for range ctx.Done() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		cancel()
		if err := srv.Shutdown(ctx); err != nil {
			klog.Errorf("Server: program exit, server exit")
		}
		wg.Done()
	}
}

func (c *CommunicateEdge) Client() {
	// TODO read map no lock protect
	selfesult := data.Result.GetLocalResultData(common.NodeIP)
	if selfesult != nil {
		tempCommunicateData := data.Result.CopyLocalResultData(common.NodeIP)
		wg := sync.WaitGroup{}
		wg.Add(len(tempCommunicateData))
		for k := range tempCommunicateData { //send to
			desNodeIP := k
			go func(wg *sync.WaitGroup) {
				if desNodeIP != common.NodeIP {
					for i := 0; i < c.CommunicateRetryTime; i++ {
						u := data.CommunicateData{SourceIP: common.NodeIP, ResultDetail: tempCommunicateData}
						if hmac, err := util.GenerateHmac(u); err != nil {
							klog.Errorf("Communicate Client: generateHmac err: %v", err)
						} else {
							u.Hmac = hmac
						}
						klog.V(4).Infof("Communicate Client: ready to put data: %v to %s", u, desNodeIP)
						requestByte, _ := json.Marshal(u)
						requestReader := bytes.NewReader(requestByte)
						ok := func() bool {
							client := http.Client{Timeout: 5 * time.Second}
							// cause peer edge-health no longer use host network, we need get peer pod ip for communication
							desPodIP, err := check.PodManager.GetPodIPByNodeIP(desNodeIP)
							if err != nil {
								klog.ErrorS(err, "get Peer Pod IP failed")
								return false
							}
							req, err := http.NewRequest("PUT", "http://"+desPodIP+":"+strconv.Itoa(c.CommunicateServerPort)+"/result", requestReader)
							if err != nil {
								klog.Errorf("Communicate Client: desPodIP %s, NewRequest err: %s", desPodIP, err.Error())
								return false
							}

							res, err := client.Do(req)
							if err != nil {
								klog.Errorf("Communicate Client: communicate to desPodIP %s failed %v", desPodIP, err)
								return false
							}
							defer func() {
								if res != nil {
									res.Body.Close()
								}
							}()
							if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
								klog.Errorf("io copy err: %s", err.Error())
							}
							if res.StatusCode != http.StatusOK {
								klog.Errorf("Communicate Client: httpResponse.StatusCode!=200, is %d", res.StatusCode)
								return false
							}

							klog.V(4).Infof("Communicate Client: put to %v status: %v succeed", desNodeIP, u)
							return true
						}()
						if ok {
							break
						}
					}
				}
				wg.Done()
			}(&wg)
		}
		wg.Wait()
	}
}

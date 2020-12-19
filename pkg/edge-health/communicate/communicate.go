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
	log "k8s.io/klog"
	"net/http"
	"strconv"
	"superedge/pkg/edge-health/common"
	"superedge/pkg/edge-health/data"
	"superedge/pkg/edge-health/util"
	pkgutil "superedge/pkg/util"
	"sync"
	"time"
)

type Communicate interface {
	Server(ctx context.Context, wg *sync.WaitGroup)
	Client()
	GetPeriod() int
}

func NewCommunicateEdge(communicatePeriod, communicateTimetout, communicateRetryTime, communicateServerPort int) Communicate {
	return CommunicateEdge{
		communicatePeriod,
		communicateTimetout,
		communicateRetryTime,
		communicateServerPort,
	}
}

type CommunicateEdge struct {
	CommunicatePeriod     int
	CommunicateTimeout    int
	CommunicateRetryTime  int
	CommunicateServerPort int
}

func (c CommunicateEdge) GetPeriod() int {
	return c.CommunicatePeriod
}

//TODO: 监听端口可变
func (c CommunicateEdge) Server(ctx context.Context, wg *sync.WaitGroup) {
	srv := &http.Server{Addr: ":" + strconv.Itoa(c.CommunicateServerPort)}
	srv.ReadTimeout = time.Duration(c.CommunicateTimeout) * time.Second
	srv.WriteTimeout = time.Duration(c.CommunicateTimeout) * time.Second
	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		var communicatedata data.CommunicateData
		if r.Body == nil {
			http.Error(w, "Please send a request body", 401)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&communicatedata)
		if err != nil {
			http.Error(w, err.Error(), 402)
			return
		}

		log.V(4).Infof("Communicate Server: received data from %v : %v", communicatedata.SourceIP, communicatedata.ResultDetail)
		if _, err := io.WriteString(w, "Received!\n"); err != nil {
			log.Errorf("Communicate Server: send response err : %v", err)
		}
		if hmac, err := util.GenerateHmac(communicatedata); err != nil {
			log.Errorf("Communicate Server: server GenerateHmac err: %v", err)
			return
		} else {
			if hmac != communicatedata.Hmac {
				log.Errorf("Communicate Server: Hmac not equal, hmac is %s, communicatedata.Hmac is %s", hmac, communicatedata.Hmac)
				http.Error(w, "Hmac not match", 403)
				return
			}
		}
		log.V(4).Infof("Communicate Server: Hmac match")

		data.Result.SetResult(&communicatedata)
		log.V(4).Infof("After communicate, result is %v", data.Result.Result)
	})

	http.HandleFunc("/debug/flags/v", pkgutil.UpdateLogLevel)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server: exit with error: %v", err)
		}
	}()

	for range ctx.Done() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Errorf("Server: program exit, server exit")
		}
		wg.Done()
	}
}

func (c CommunicateEdge) Client() {
	if _, ok := data.Result.Result[common.LocalIp]; ok {
		tempCommunicateData := data.Result.CopyLocalResultData(common.LocalIp)
		wg := sync.WaitGroup{}
		wg.Add(len(tempCommunicateData))
		for k := range tempCommunicateData { //send to
			des := k
			go func(wg *sync.WaitGroup) {
				if des != common.LocalIp {
					for i := 0; i < c.CommunicateRetryTime; i++ {
						u := data.CommunicateData{SourceIP: common.LocalIp, ResultDetail: tempCommunicateData}
						if hmac, err := util.GenerateHmac(u); err != nil {
							log.Errorf("Communicate Client: generateHmac err: %v", err)
						} else {
							u.Hmac = hmac
						}
						log.V(4).Infof("Communicate Client: ready to put data: %v to %s", u, des)
						requestByte, _ := json.Marshal(u)
						requestReader := bytes.NewReader(requestByte)
						ok := func() bool {
							client := http.Client{Timeout: time.Duration(c.CommunicateTimeout) * time.Second}
							req, err := http.NewRequest("PUT", "http://"+des+":"+strconv.Itoa(c.CommunicateServerPort)+"/result", requestReader)
							if err != nil {
								log.Errorf("Communicate Client: %s, NewRequest err: %s", des, err.Error())
								return false
							}

							res, err := client.Do(req)
							if err != nil {
								log.Errorf("Communicate Client: communicate to %v failed %v", des, err)
								return false
							}
							defer func() {
								if res != nil {
									res.Body.Close()
								}
							}()
							if _, err := io.Copy(ioutil.Discard, res.Body); err != nil {
								log.Errorf("io copy err: %s", err.Error())
							}
							if res.StatusCode != http.StatusOK {
								log.Errorf("Communicate Client: httpResponse.StatusCode!=200, is %d", res.StatusCode)
								return false
							}

							log.V(4).Infof("Communicate Client: put to %v status: %v succeed", des, u)
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

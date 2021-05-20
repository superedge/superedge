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

package proxy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"k8s.io/klog"

	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/transport"
)

type holder struct {
	key        string
	isStart    bool
	request    *http.Request
	syncTime   time.Duration
	recordTime time.Time

	peerDuration   time.Duration
	expireDuration time.Duration

	ticker *time.Ticker

	requestCh chan<- *http.Request
	stopCh    chan struct{}
}

func newHolder(r *http.Request, key string, syncTime time.Duration, ch chan<- *http.Request) *holder {
	h := &holder{
		request:    r,
		key:        key,
		syncTime:   syncTime,
		recordTime: time.Now(),
		requestCh:  ch,
		stopCh:     make(chan struct{}),

		peerDuration:   30 * time.Second,
		expireDuration: 10 * time.Minute,
	}
	return h
}

func (h *holder) start() {
	if h.isStart {
		h.recordTime = time.Now()
		return
	}

	klog.V(2).Infof("Start holder %s", h.key)
	h.isStart = true
	h.ticker = time.NewTicker(h.syncTime)
	go h.run()
}

func (h *holder) close() {
	klog.V(4).Infof("Close holder %s", h.key)
	if !h.isStart {
		return
	}

	h.isStart = false
	h.recordTime = time.Now()
	h.stopCh <- struct{}{}
}

func (h *holder) run() {
	klog.V(4).Infof("Begin to run holder %s", h.key)
	for {
		select {
		case <-h.stopCh:
			klog.Infof("Stop holder %s loop", h.key)
			return
		case <-h.ticker.C:
			h.requestCh <- h.request
		}
	}
}

func (h *holder) expired() bool {
	if h.isStart {
		return time.Now().After(h.recordTime.Add(h.expireDuration))
	} else {
		return time.Now().After(h.recordTime.Add(h.peerDuration))
	}
}

// RequestCacheController caches all 'get request' and 'watch request' info
type RequestCacheController struct {
	listRequestMap  map[string]*holder
	watchRequestMap map[string]*http.Request

	syncTime time.Duration
	url      string

	lock sync.Mutex

	requestCh        chan *http.Request
	transportManager *transport.TransportManager
}

func NewRequestCacheController(config *config.LiteServerConfig, transportManager *transport.TransportManager) *RequestCacheController {
	c := &RequestCacheController{
		listRequestMap:   make(map[string]*holder),
		watchRequestMap:  make(map[string]*http.Request),
		syncTime:         config.SyncDuration,
		url:              fmt.Sprintf("https://127.0.0.1:%d", config.Port),
		transportManager: transportManager,
	}
	c.requestCh = make(chan *http.Request, 100)
	return c
}

func (c *RequestCacheController) Run(stopCh <-chan struct{}) {
	klog.Infof("Request cache controller begin run")
	go c.runGC(stopCh)
	for {
		select {
		case <-stopCh:
			klog.Infof("Receive stop channel, exit request cache controller")
			return
		case r := <-c.requestCh:
			klog.V(2).Infof("Update list request, url %s", r.URL.String())
			go c.doRequest(r)
		}
	}
}

// doRequest sends list resource request to lite-apiserver and thus syncs list resource response cache
func (c *RequestCacheController) doRequest(r *http.Request) {
	var commonName string
	if r.TLS != nil {
		for _, cert := range r.TLS.PeerCertificates {
			if !cert.IsCA {
				commonName = cert.Subject.CommonName
				break
			}
		}
	}

	tr := c.transportManager.GetTransport(commonName)
	client := http.Client{
		Transport: tr.Transport,
	}

	newReq, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", c.url, r.URL.String()), bytes.NewReader([]byte{}))
	if err != nil {
		klog.Errorf("parse path error %v", err)
		return
	}
	CopyHeader(newReq.Header, r.Header)
	newReq.Header.Set(EdgeUpdateHeader, time.Now().String())
	defer newReq.Body.Close()

	resp, err := client.Do(newReq)
	if err != nil {
		klog.Errorf("auto update request do err %v", err)
		return
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		klog.Errorf("io copy err: %s", err.Error())
	}
}

func (c *RequestCacheController) runGC(stopCh <-chan struct{}) {
	watchGCTicker := time.NewTicker(time.Second)
	listGCTicker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-stopCh:
			klog.Infof("receive stop channel, exit request gc controller")
			return
		case <-listGCTicker.C:
			c.lock.Lock()
			for k, h := range c.listRequestMap {
				if _, e := c.watchRequestMap[k]; !e {
					if h.expired() {
						klog.Infof("request key %s, url %s has expired, delete it", k, h.request.URL.Path)
						h.close()
						delete(c.listRequestMap, k)
					}
				}
			}
			c.lock.Unlock()
		case <-watchGCTicker.C:
			c.lock.Lock()
			for k, r := range c.watchRequestMap {
				req := r
				select {
				case <-req.Context().Done():
					klog.V(4).Infof("Watch %s connection closed.", k)
					if h, e := c.listRequestMap[k]; e {
						h.close()
					}
					delete(c.watchRequestMap, k)
				default:
				}
			}
			c.lock.Unlock()
		}
	}
}

// Get:
// Path /api/v1/services
// RawQuery limit=500&resourceVersion=0
//
// Watch:
// Path /api/v1/services
// RawQuery allowWatchBookmarks=true&resourceVersion=1886882&timeout=8m1s&timeoutSeconds=481&watch=true
//
// we check request pair by url.path
func (c *RequestCacheController) AddRequest(r *http.Request, userAgent string, list bool, watch bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if watch {
		klog.V(4).Infof("Receive watch request (%s, %s)", userAgent, r.URL)
		c.addWatchRequest(r, userAgent)
		return
	}

	if list {
		klog.V(4).Infof("Receive list request (%s, %s)", userAgent, r.URL)
		c.addListRequest(r, userAgent)
		return
	}
}

func (c *RequestCacheController) DeleteRequest(req *http.Request, userAgent string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := c.key(userAgent, req.URL.Path)
	klog.Infof("Receive delete watch request %s. Stop list in background", key)

	if h, e := c.listRequestMap[key]; e {
		h.close()
	}
	delete(c.listRequestMap, key)
	delete(c.watchRequestMap, key)
}

func (c *RequestCacheController) addListRequest(req *http.Request, userAgent string) {
	key := c.key(userAgent, req.URL.Path)

	if h, e := c.listRequestMap[key]; e {
		klog.Infof("Delete old list request %s", key)
		h.close()
	}

	klog.V(2).Infof("Create or update list request %s", key)
	c.listRequestMap[key] = newHolder(req, key, c.syncTime, c.requestCh)
}

func (c *RequestCacheController) addWatchRequest(req *http.Request, userAgent string) {
	key := c.key(userAgent, req.URL.Path)

	h, e := c.listRequestMap[key]
	if !e {
		klog.Infof("Only watch request, ignore it %s", key)
		return
	}

	h.start()

	klog.V(2).Infof("Create or update watch request %s", key)
	c.watchRequestMap[key] = req
}

func (c *RequestCacheController) key(userAgent string, path string) string {
	return fmt.Sprintf("%s_%s", userAgent, path)
}

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
	"io"
	"net/http"

	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog/v2"
)

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			klog.V(6).Infof("Copy header for key=%s, value=%s", k, v)
			dst.Add(k, v)
		}
	}
}

// WithRequestAccept delete header Accept, use default application/json
func WithRequestAccept(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			if info, ok := apirequest.RequestInfoFrom(req.Context()); ok {
				if info.IsResourceRequest {
					req.Header.Del("Accept")
				}
			}
		}
		handler.ServeHTTP(w, req)
	})
}

type multiWrite struct {
	read   io.ReadCloser
	writes []io.WriteCloser
}

func (w *multiWrite) Read(p []byte) (n int, err error) {

	n, err = w.Read(p)
	if err != nil {
		klog.Errorf("multiWrite failed to read source data, error: %v", err)
		return n, err
	}
	for i := 0; i < len(w.writes); i++ {
		_, err = w.writes[i].Write(p)
		if err != nil {
			klog.Errorf("multiWrite failed to write data to the pipe, error: %v")
			return n, err
		}
	}
	return n, err
}

func (w *multiWrite) Close() error {

	for k := 0; k < len(w.writes); k++ {
		err := w.writes[k].Close()
		if err != nil {
			klog.Errorf("multiWrite failed to close PipeWriter, error: %v", err)
			return err
		}
	}
	return nil
}

func MultiWrite(read io.ReadCloser, number int) []io.ReadCloser {
	if number < 2 {
		return nil
	}
	rs := make([]io.ReadCloser, number)
	ws := make([]io.WriteCloser, number-1)
	for i := 1; i < number; i++ {
		r, w := io.Pipe()
		rs[i] = r
		ws[i-1] = w
	}
	rs[0] = &multiWrite{read, ws}
	return rs
}

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
	"strings"

	"github.com/superedge/superedge/pkg/lite-apiserver/constant"
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

// WithRequestAccept replace header Accept, use default application/json
func WithRequestAccept(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			if info, ok := apirequest.RequestInfoFrom(req.Context()); ok {
				if info.IsResourceRequest {
					// don't del Accept, just replace protobuf to json
					newAccept := strings.ReplaceAll(req.Header.Get("Accept"), constant.Protobuf, constant.Json)
					if newAccept != "" {
						req.Header.Set("Accept", newAccept)
					}
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

//multiRead prioritizes reading of the source ReaderCloser object
func (w *multiWrite) Read(p []byte) (n int, err error) {
	n, err = w.read.Read(p)
	if err != nil {
		if err != io.EOF {
			w.Close()
			return n, err
		}
	}
	if n > 0 {
		for i := 0; i < len(w.writes); i++ {
			_, pipeErr := w.writes[i].Write(p[:n])
			if pipeErr != nil {
				klog.Errorf("multiWrite failed to write data to the pipe, error: %v", err)
				//In order not to affect the reading of the source ReaderCloser, close the writecloser object that failed to write
				w.writes[i].Close()
			}
		}
	}
	return
}

func (w *multiWrite) Close() error {
	err := w.read.Close()
	if err != nil {
		klog.Errorf("multiWrite failed to close source read, error: %v", err)
		return err
	}

	for k := 0; k < len(w.writes); k++ {
		err = w.writes[k].Close()
		if err != nil {
			klog.Errorf("multiWrite failed to close PipeWriter, error: %v", err)
			return err
		}
	}
	return nil
}

// MultiWrite The ReadCloser object in the returned array needs to use multiple goroutines to read at the same time
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
	//source read
	rs[0] = &multiWrite{read, ws}

	return rs
}

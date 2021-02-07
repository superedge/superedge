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
	"fmt"
	"io"

	"net/http"

	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog"
)

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			klog.V(6).Infof("Copy header for auto update: key=%s, value=%s", k, v)
			dst.Add(k, v)
		}
	}
}

// WithRequestAccept delete Accept header is set, application/json will be used
func WithRequestAccept(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			if info, ok := apirequest.RequestInfoFrom(req.Context()); ok {
				if info.IsResourceRequest {
					req.Header.Del("Accept")
				}
			} else {
				klog.Errorf("!!!not ok!!!")
			}
		}
		handler.ServeHTTP(w, req)
	})
}

// NewDupReadCloser create an dupReadCloser object
func NewDupReadCloser(rc io.ReadCloser) (io.ReadCloser, io.ReadCloser) {
	pr, pw := io.Pipe()
	dr := &dupReadCloser{
		rc: rc,
		pw: pw,
	}

	return dr, pr
}

type dupReadCloser struct {
	rc io.ReadCloser
	pw *io.PipeWriter
}

// Read read data into p and write into pipe
func (dr *dupReadCloser) Read(p []byte) (n int, err error) {
	n, err = dr.rc.Read(p)
	if n > 0 {
		if n, err := dr.pw.Write(p[:n]); err != nil {
			klog.Errorf("dualReader: failed to write %v", err)
			return n, err
		}
	}

	return
}

// Close close dupReader
func (dr *dupReadCloser) Close() error {
	errs := make([]error, 0)
	if err := dr.rc.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := dr.pw.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return fmt.Errorf("failed to close dupReader, %v", errs)
	}

	return nil
}

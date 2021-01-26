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

package util

import (
	"fmt"
	"github.com/spf13/pflag"
	"io/ioutil"
	"net/http"

	"k8s.io/klog/v2"
)

// PrintFlags logs the flags in the flagset
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}

func UpdateLogLevel(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "PUT":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writePlainText(http.StatusBadRequest, "error reading request body: "+err.Error(), w)
			return
		}
		defer r.Body.Close()
		response, err := updateLogLevel(string(body))
		if err != nil {
			writePlainText(http.StatusBadRequest, err.Error(), w)
			return
		}
		writePlainText(http.StatusOK, response, w)
		return
	default:
		writePlainText(http.StatusNotAcceptable, "unsupported http method", w)
		return
	}
}

func updateLogLevel(val string) (string, error) {
	var level klog.Level
	if err := level.Set(val); err != nil {
		return "", fmt.Errorf("failed set klog.logging.verbosity %s: %v", val, err)
	}
	return "successfully set klog.logging.verbosity to " + val, nil
}

// writePlainText renders a simple string response.
func writePlainText(statusCode int, text string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, text)
}

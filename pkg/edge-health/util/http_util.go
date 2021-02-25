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
	"io"
	"io/ioutil"
	"net/http"
)

// DoRequestAndDiscard send http request using client param and discard the corresponding response.
func DoRequestAndDiscard(client *http.Client, req *http.Request) error {
	// Use httpClient to send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	// Close the connection to reuse it
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request %+v failed, StatusCode is %d", req, resp.StatusCode)
	}
	// Discard resp body
	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return fmt.Errorf("Discard response body err %s", err.Error())
	}
	return nil
}

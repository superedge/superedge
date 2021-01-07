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

package main

import (
	"flag"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"os"
	"testing"
)

func init() {
	flag.String("m", "cloud", "mode")
	flag.String("c", "../../conf/cloud_mode.toml", "config")
}

func TestMain(m *testing.M) {
	flag.Parse()
	m.Run()
	os.Exit(0)
}
func Test_Server(t *testing.T) {
	main()
}

func Test_Client(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	main()
}

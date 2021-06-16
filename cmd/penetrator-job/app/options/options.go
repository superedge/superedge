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

package options

import (
	"flag"
	"github.com/spf13/pflag"
)

type Options struct {
	SecretPath  string
	JobConfPath string
	Nodes       int
}

func NewJobOptions() *Options {
	return &Options{
		SecretPath:  "/etc/superedge/penetrator/job/secret/",
		JobConfPath: "/etc/superedge/penetrator/job/conf/",
		Nodes:       5,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SecretPath, "secret-path", o.SecretPath, "Specify the path of sshkey and password")
	fs.StringVar(&o.JobConfPath, "jobconf-path", o.JobConfPath, "Specify the path of the job configuration file")
	flag.IntVar(&o.Nodes, "nodes", o.Nodes, "The number of concurrently installed nodes")
}

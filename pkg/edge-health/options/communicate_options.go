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

import "github.com/spf13/pflag"

type CommunOptions struct {
	CommunicatePeriod     int
	CommunicateTimeout    int
	CommunicateRetryTime  int
	CommunicateServerPort int
}

func NewCommunOptions() *CommunOptions {
	return &CommunOptions{}
}

func (o *CommunOptions) Default() {
	o.CommunicatePeriod = 60
	o.CommunicateRetryTime = 1
	o.CommunicateTimeout = 2
	o.CommunicateServerPort = 51005
}

func (o *CommunOptions) Validate() []error {
	return nil
}

func (o *CommunOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.IntVar(&o.CommunicatePeriod, "communicateperiod", o.CommunicatePeriod, "communicate period")
	fs.IntVar(&o.CommunicateTimeout, "communicatetimetout", o.CommunicateTimeout, "communicate timetout")
	fs.IntVar(&o.CommunicateRetryTime, "communicateretrytime", o.CommunicateRetryTime, "communicate retry time")
	fs.IntVar(&o.CommunicateServerPort, "communicateserverport", o.CommunicateServerPort, "communicate server port")
}

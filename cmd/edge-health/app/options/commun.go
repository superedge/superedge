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
	"fmt"
	"github.com/spf13/pflag"
)

type CommunOptions struct {
	CommunPeriod     int
	CommunTimeout    int
	CommunRetries    int
	CommunServerPort int
}

func NewCommunOptions() *CommunOptions {
	return &CommunOptions{}
}

func (o *CommunOptions) Default() {
	o.CommunPeriod = 10
	o.CommunRetries = 1
	o.CommunTimeout = 3
	o.CommunServerPort = 51005
}

func (o *CommunOptions) Validate() []error {
	var errs []error
	if o.CommunPeriod < 0 {
		errs = append(errs, fmt.Errorf("Invalid communicate period %d", o.CommunPeriod))
	}
	if o.CommunTimeout < 0 {
		errs = append(errs, fmt.Errorf("Invalid communicate timeout %d", o.CommunTimeout))
	}
	if o.CommunRetries < 0 {
		errs = append(errs, fmt.Errorf("Invalid communicate retry times %d", o.CommunRetries))
	}
	return errs
}

func (o *CommunOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.IntVar(&o.CommunPeriod, "commun-period", o.CommunPeriod, "Period of communicate")
	fs.IntVar(&o.CommunTimeout, "commun-timetout", o.CommunTimeout, "Communicate timeout")
	fs.IntVar(&o.CommunRetries, "commun-retries", o.CommunRetries, "Communicate retry times")
	fs.IntVar(&o.CommunServerPort, "commun-server-port", o.CommunServerPort, "Communicate server port")
}

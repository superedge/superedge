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

type VoteOptions struct {
	VotePeriod  int
	VoteTimeout int // Vote will be timeout after VoteTimeout
}

func NewVoteOptions() *VoteOptions {
	return &VoteOptions{}
}

func (o *VoteOptions) Default() {
	o.VotePeriod = 10
	o.VoteTimeout = 60
}

func (o *VoteOptions) Validate() []error {
	var errs []error
	if o.VotePeriod <= 0 {
		errs = append(errs, fmt.Errorf("Invalid vote period %d", o.VotePeriod))
	}
	if o.VoteTimeout <= 0 {
		errs = append(errs, fmt.Errorf("Invalid vote timeout %d", o.VoteTimeout))
	}
	return errs
}

func (o *VoteOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.IntVar(&o.VotePeriod, "vote-period", o.VotePeriod, "Period of vote")
	fs.IntVar(&o.VoteTimeout, "vote-timeout", o.VoteTimeout, "Vote timeout")
}

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

package daemon

import (
	"github.com/superedge/superedge/pkg/edge-health/checkplugin"
	"github.com/superedge/superedge/pkg/edge-health/commun"
	"github.com/superedge/superedge/pkg/edge-health/config"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"github.com/superedge/superedge/pkg/edge-health/vote"
)

type EdgeHealthDaemon struct {
	cfg         *config.EdgeHealthConfig
	metadata    *metadata.EdgeHealthMetadata
	checkPlugin checkplugin.Plugin
}

func NewEdgeHealthDaemon(c *config.EdgeHealthConfig) *EdgeHealthDaemon {
	return &EdgeHealthDaemon{
		cfg:         c,
		metadata:    metadata.NewEdgeHealthMetadata(),
		checkPlugin: checkplugin.NewPlugin(),
	}
}

func (ehd *EdgeHealthDaemon) Run(stopCh <-chan struct{}) {
	// Execute edge health prepare and check
	go ehd.PrepareAndCheck(stopCh)

	// Execute communication
	communEdge := commun.NewCommunEdge(&ehd.cfg.Commun)
	go communEdge.Commun(ehd.metadata.CheckMetadata, ehd.cfg.ConfigMapInformer, ehd.cfg.Node.LocalIp, stopCh)

	// Execute vote
	vote := vote.NewVoteEdge(&ehd.cfg.Vote)
	go vote.Vote(ehd.metadata, ehd.cfg.Kubeclient, ehd.cfg.Node.LocalIp, stopCh)

	<-stopCh
}

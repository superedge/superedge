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

package join

import (
	"github.com/spf13/cobra"

	"github.com/superedge/superedge/pkg/edgeadm/cmd"
)

/*
添加master的一些注意事项：
- 需要把添加maste的hostname，写入第一个master的/etc/host: masterIP hostname, 这样在其他master节点才能用其他master的hostname访问
*/

func NewJoinCMD() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "join",
		Short: "Output edgeadm build info",
	}

	cmds.AddCommand(cmd.NewVersionCMD()) // example，Please implement specific command logic

	return cmds
}

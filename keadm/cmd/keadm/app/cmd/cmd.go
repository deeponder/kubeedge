/*
Copyright 2019 The KubeEdge Authors.

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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/beta"
	cloud "github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/cloud"
	"github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/debug"
	edge "github.com/kubeedge/kubeedge/keadm/cmd/keadm/app/cmd/edge"
)

var (
	keadmLongDescription = `
    +----------------------------------------------------------+
    | KEADM                                                    |
    | Easily bootstrap a KubeEdge cluster                      |
    |                                                          |
    | Please give us feedback at:                              |
    | https://github.com/kubeedge/kubeedge/issues              |
    +----------------------------------------------------------+

    Create a cluster with cloud node
    (which controls the edge node cluster), and edge nodes
    (where native containerized application, in the form of
    pods and deployments run), connects to devices.

`
	keadmExample = `
    +----------------------------------------------------------+
    | On the cloud machine:                                    |
    +----------------------------------------------------------+
    | master node (on the cloud)# sudo keadm init              |
    +----------------------------------------------------------+

    +----------------------------------------------------------+
    | On the edge machine:                                     |
    +----------------------------------------------------------+
    | worker node (at the edge)# sudo keadm join <flags>       |
    +----------------------------------------------------------+

    You can then repeat the second step on, as many other machines as you like.
`
)

/*
keadm支持的命令定义，包括：
1. init。
2. join
3. reset
4. version
5. gettoken
6. debug
7. beta


*/

// NewKubeedgeCommand returns cobra.Command to run keadm commands
func NewKubeedgeCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:     "keadm",
		Short:   "keadm: Bootstrap KubeEdge cluster",
		Long:    keadmLongDescription,
		Example: keadmExample,
	}

	cmds.ResetFlags()
	cmds.AddCommand(cloud.NewCloudInit())
	cmds.AddCommand(edge.NewEdgeJoin())
	cmds.AddCommand(NewKubeEdgeReset())
	cmds.AddCommand(NewCmdVersion())
	cmds.AddCommand(cloud.NewGettoken())
	cmds.AddCommand(debug.NewEdgeDebug())
	cmds.AddCommand(beta.NewBeta())

	return cmds
}

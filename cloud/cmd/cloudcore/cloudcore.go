package main

import (
	"os"

	"k8s.io/component-base/logs"

	"github.com/kubeedge/kubeedge/cloud/cmd/cloudcore/app"
)

// cloudcore各个模块的启动入口， 也是借助cobra
// 整个cloudCore是一个进程，内部有cloudhub等多个模块
func main() {
	command := app.NewCloudCoreCommand()
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

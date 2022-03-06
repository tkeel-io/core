package main

import (
	"context"

	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/runtime/mock"
	"github.com/tkeel-io/kit/log"
)

func main() {
	stopCh := make(chan struct{}, 0)
	placement.Initialize()
	log.InitLogger("core.node", "DEBUG", true)
	node := runtime.NewNode(context.Background(), mock.NewRepo(), mock.NewDispatcher())

	err := node.Start(runtime.NodeConf{
		Sources: []string{
			"kafka://139.198.125.147:9092/core0/core",
			"kafka://139.198.125.147:9092/core1/core",
		},
	})

	if nil != err {
		panic(err)
	}

	<-stopCh
}

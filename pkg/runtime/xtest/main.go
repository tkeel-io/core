package main

import (
	"context"
	"time"

	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/runtime/mock"
	"github.com/tkeel-io/kit/log"
)

func main() {
	placement.Initialize()
	log.InitLogger("core.node", "DEBUG", true)
	node := runtime.NewNode(context.Background(), mock.NewRepo(), mock.NewDispatcher())

	err := node.Start(runtime.NodeConf{
		Sources: []string{
			"kafka://192.168.0.103:9092/core/core",
		},
	})

	if nil != err {
		panic(err)
	}

	time.Sleep(time.Hour)
}

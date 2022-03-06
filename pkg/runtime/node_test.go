package runtime

import (
	"context"
	"net/url"
	"testing"

	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/runtime/mock"
	"github.com/tkeel-io/kit/log"
)

func TestNode_Start(t *testing.T) {
	stopCh := make(chan struct{})
	placement.Initialize()
	log.InitLogger("core.node", "DEBUG", true)
	node := NewNode(context.Background(), nil, mock.NewDispatcher())

	err := node.Start([]string{
		"kafka://139.198.125.147:9092/core/core",
	},
	)

	if nil != err {
		panic(err)
	}

	<-stopCh
}

func TestParse(t *testing.T) {
	urlText := "partition://admin:admin@192.168.12.1;192.168.12.1/core/0"
	URL, _ := url.Parse(urlText)
	t.Log(URL)
}

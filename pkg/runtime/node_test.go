package runtime

import (
	"context"
	"net/url"
	"testing"

	"github.com/tkeel-io/core/pkg/runtime/mock"
	"github.com/tkeel-io/kit/log"
)

func TestNode_Start(t *testing.T) {
	log.InitLogger("core.node", "DEBUG", true)
	node := NewNode(context.Background(), mock.NewRepo(), mock.NewDispatcher())

	err := node.Start(NodeConf{
		Sources: []string{
			"kafka://192.168.0.103:9092/core/core",
		},
	})

	if nil != err {
		panic(err)
	}
}

func TestParse(t *testing.T) {
	urlText := "partition://admin:admin@192.168.12.1;192.168.12.1/core/0"
	URL, _ := url.Parse(urlText)
	t.Log(URL)
}

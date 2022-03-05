package runtime

import (
	"context"
	"net/url"
	"testing"

	"github.com/tkeel-io/core/pkg/runtime/mock"
)

func TestNode_Start(t *testing.T) {
	node := NewNode(context.TODO(), mock.NewRepo(), nil)
	node.Start(NodeConf{
		Source: SourceConf{
			Topics:    []string{"core"},
			Brokers:   []string{"192.168.0.103:9092"},
			GroupName: "core",
		},
	})
}

func TestParse(t *testing.T) {
	urlText := "partition://admin:admin@192.168.12.1;192.168.12.1/core/0"
	URL, _ := url.Parse(urlText)
	t.Log(URL)
}

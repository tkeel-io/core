package runtime3

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/tkeel-io/core/pkg/runtime3/mock"
)

func TestRuntime_Start(t *testing.T) {
	rt := NewRuntime(context.TODO(), mock.NewRepo(), nil)
	rt.Start(RuntimeConfig{
		Source: SourceConf{
			Topics:    []string{"core"},
			Brokers:   []string{"192.168.0.103:9092"},
			GroupName: "core",
		},
	})

	time.Sleep(1 * time.Hour)
}

func TestParse(t *testing.T) {
	urlText := "partition://admin:admin@192.168.12.1;192.168.12.1/core/0"
	URL, _ := url.Parse(urlText)
	t.Log(URL)
}

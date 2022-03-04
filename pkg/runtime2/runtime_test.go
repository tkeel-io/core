package runtime2

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestRuntime_Start(t *testing.T) {
	rt := NewRuntime(context.TODO(), Dao{}, Dispatch{})
	rt.Start(RuntimeConfig{
		Source: SourceConf{
			Topic:      "core",
			Brokers:    []string{"192.168.0.103:9092"},
			Partitions: []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
	})

	time.Sleep(1 * time.Hour)
}

func TestParse(t *testing.T) {
	urlText := "partition://admin:admin@192.168.12.1;192.168.12.1/core/0"
	URL, _ := url.Parse(urlText)
	t.Log(URL)
}

package runtime2

import (
	"net/url"
	"testing"
)

func TestRuntime_Start(t *testing.T) {

	rt := NewRuntime(Dao{}, Dispatch{})
	rt.Start(RuntimeConfig{
		Source: SourceConf{
			Topic:      "core",
			Brokers:    []string{"192.168.0.103:9092"},
			Partitions: []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
	})

}

func TestParse(t *testing.T) {
	urlText := "partition://admin:admin@192.168.12.1;192.168.12.1/core/0"
	URL, _ := url.Parse(urlText)
	t.Log(URL)
}

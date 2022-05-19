package clickhouse

import (
	"context"
	"math/rand"
	"testing"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"
)

func Test_clickhouse_Write(t *testing.T) {
	c := newClickhouse()
	err := c.Init(resource.Metadata{
		Name: "clickhouse",
		Properties: map[string]interface{}{
			"urls":     []string{"clickhouse://default:C1ickh0use@clickhouse-tkeel-core:9000"},
			"table":    "test15",
			"database": "core",
		},
	})
	t.Log(err)
	if err != nil {
		return
	}
	for i := 0; i < 10000; i++ {
		req := &tseries.TSeriesRequest{
			Data: []*tseries.TSeriesData{
				{
					Measurement: "",
					Tags: map[string]string{
						"id": "iotd-123",
					},
					Fields: map[string]float32{
						"abcd": rand.Float32(),
						"abc":  rand.Float32(),
					},
					Timestamp: time.Now().UnixNano(),
				},
			},
			Metadata: map[string]string{},
		}
		go c.Write(context.Background(), req)
		t.Log(i)
	}
	time.Sleep(time.Second * 10)
}

func Test_clickhouse_Query(t *testing.T) {
	c := newClickhouse()
	err := c.Init(resource.Metadata{
		Name: "clickhouse",
		Properties: map[string]interface{}{
			"urls":     []string{"clickhouse://default:C1ickh0use@clickhouse-tkeel-core:9000"},
			"table":    "test5",
			"database": "core",
		},
	})
	t.Log(err)
	if err != nil {
		return
	}
	resp, err := c.Query(context.Background(), &pb.GetTSDataRequest{
		Id:          "iotd-123",
		StartTime:   time.Now().Unix() - 500,
		EndTime:     time.Now().Unix(),
		Identifiers: "abc,abcd",
		PageNum:     1,
		PageSize:    3,
	})
	t.Log(err)
	t.Log(resp)
}

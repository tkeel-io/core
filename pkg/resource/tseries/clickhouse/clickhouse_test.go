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

func Test_clickhouse_Init(t *testing.T) {
	c := newClickhouse()
	err := c.Init(resource.Metadata{
		Name: "clickhouse",
		Properties: map[string]interface{}{
			"host":     "127.0.0.1",
			"port":     "9000",
			"db_name":  "core2",
			"table":    "test5",
			"user":     "default",
			"password": "tkeel123!",
		},
	})
	t.Log(err)
	if err != nil {
		return
	}
	for i := 0; i < 100; i++ {
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
					Timestamp: time.Now().UnixMilli(),
				},
			},
			Metadata: map[string]string{},
		}
		c.Write(context.Background(), req)
		t.Log(i)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)))
	}
	time.Sleep(time.Second)
}

func Test_clickhouse_Query(t *testing.T) {
	c := newClickhouse()
	err := c.Init(resource.Metadata{
		Name: "clickhouse",
		Properties: map[string]interface{}{
			"host":     "127.0.0.1",
			"port":     "9000",
			"db_name":  "core2",
			"table":    "test5",
			"user":     "default",
			"password": "tkeel123!",
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

package clickhouse

import (
	"context"
	"testing"
	"time"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
)

func TestClickhouse_genSql(t *testing.T) {
	c := Clickhouse{
		option: &Option{
			Urls:   []string{"tcp://139.19.1.173:9089?username=default&password=qingcloud2019&database=iot_manage_dev&alt_hosts=139.198.18.173:9090,139.198.18.173:9091"},
			DbName: "dbname",
			Table:  "table",
			Fields: map[string]Field{},
		},
		balance: nil,
	}
	row := execNode{
		ts:     0,
		fields: []string{"a", "b", "c"},
		args:   []interface{}{"1", "2", "3"},
	}
	sql := c.genSql(&row)
	t.Log(sql)
}

func TestClickhouse(t *testing.T) {
	metadata := resource.Metadata{
		Name: "myck",
		Properties: map[string]interface{}{
			"database": "core",
			"urls":     []interface{}{"http://default:tkeel123!@139.198.112.150:8123"},
			"table":    "event_data1",
		},
	}
	ck := NewClickhouse()
	ck.Init(metadata)
	req := rawdata.RawDataRequest{
		Data:     []*rawdata.RawData{},
		Metadata: map[string]string{},
	}
	req.Metadata["type"] = "device"
	req.Metadata["path"] = "abc2"
	req.Data = append(req.Data, &rawdata.RawData{
		EntityID:  "iotd-124",
		Path:      "adcd",
		Values:    "ddddd",
		Timestamp: time.Now(),
	})
	err := ck.Write(context.Background(), &req)
	t.Log(err)
	//ck.Query()
}

func TestClickhouse_Query(t *testing.T) {
	metadata := resource.Metadata{
		Name: "myck",
		Properties: map[string]interface{}{
			"database": "core1",
			"urls":     []interface{}{"http://default:tkeel@139.19.112.15:8123"},
			"table":    "event_data",
		},
	}
	ck := NewClickhouse()
	ck.Init(metadata)
	req := &pb.GetRawdataRequest{
		EntityId:     "iotd-124",
		StartTime:    time.Now().Unix() - 3600*24 + 8*3600,
		EndTime:      time.Now().Unix() + 8*3600,
		Path:         "adcd",
		PageNum:      1,
		PageSize:     4,
		IsDescending: false,
		Filters:      map[string]string{},
	}
	req.Filters["path"] = "abc1,abc2"
	resp, err := ck.Query(context.Background(), req)
	if err != nil {
		t.Log(err)
	} else {
		t.Log(resp)
	}
}

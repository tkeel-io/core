package clickhouse

import (
	"testing"
	"time"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/rawdata"
)

func TestClickhouse_genSql(t *testing.T) {
	c := Clickhouse{
		option: &Option{
			Urls:   []string{"http://default:C1ickh0use@clickhouse-my-ck:8123"},
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
			"urls":     []interface{}{"http://default:C1ickh0use@clickhouse-my-ck:8123"},
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
	//	err := ck.Write(context.Background(), &req)
	//	t.Log(err)
}

func TestClickhouse_Query(t *testing.T) {
	metadata := resource.Metadata{
		Name: "myck",
		Properties: map[string]interface{}{
			"database": "core",
			"urls":     []interface{}{"http://default:C1ickh0use@clickhouse-my-ck:8123"},
			"table":    "event_data",
		},
	}
	ck := NewClickhouse()
	err := ck.Init(metadata)
	if err != nil {
		return
	}
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
	//	resp, err := ck.Query(context.Background(), req)
	if err != nil {
		t.Log(err)
	} else {
		//		t.Log(resp)
	}
}

package rawdata

import (
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
)

var (
	json              = jsoniter.ConfigCompatibleWithStandardLibrary
	registeredRawData = make(map[string]Generator)
)

type RawData struct {
	ID        string    `db:"id"`
	EntityID  string    `db:"entity_id"`
	Path      string    `db:"path"`
	Values    string    `db:"values"`
	Timestamp time.Time `db:"timestamp"`
	Tag       []string  `db:"tag"`
}

func (r *RawData) Bytes() []byte {
	byt, err := json.Marshal(r)
	if err == nil {
		return byt
	}
	return []byte("")
}

type Request struct {
	Data     []*RawData
	Metadata map[string]string
}

type Response struct {
	Data     []*RawData
	Metadata map[string]string
}

type Service interface {
	Init(resource.Metadata) error
	Write(ctx context.Context, req *Request) error
	Query(ctx context.Context, req *pb.GetRawdataRequest) (*pb.GetRawdataResponse, error)
	GetMetrics() (count, storage, total, used float64)
}

type Generator func() Service

func NewRawDataService(name string) Service {
	if generator, has := registeredRawData[name]; has {
		return generator()
	}
	return registeredRawData["noop"]()
}

func Register(name string, handler Generator) {
	registeredRawData[name] = handler
}

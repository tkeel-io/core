package rawdata

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/resource"
)

var registeredRawData = make(map[string]RawDataGenerator)

type RawData struct {
	ID        string    `db:"id"`
	EntityID  string    `db:"entity_id"`
	Path      string    `db:"path"`
	Values    string    `db:"values"`
	Timestamp time.Time `db:"timestamp"` //nolint
	Tag       []string  `db:"tag"`
}

func (r *RawData) Bytes() []byte {
	byt, err := json.Marshal(r)
	if err == nil {
		return byt
	}
	return []byte("")
}

type RawDataRequest struct {
	Data     []*RawData
	Metadata map[string]string
}

type RawDataResponse struct {
	Data     []*RawData
	Metadata map[string]string
}

type RawDataService interface {
	Init(resource.Metadata) error
	Write(ctx context.Context, req *RawDataRequest) error
	Query(ctx context.Context, req *pb.GetRawdataRequest) (*pb.GetRawdataResponse, error)
}

type RawDataGenerator func() RawDataService

func NewRawDataService(name string) RawDataService {
	if generator, has := registeredRawData[name]; has {
		return generator()
	}
	return registeredRawData["noop"]()
}

func Register(name string, handler RawDataGenerator) {
	registeredRawData[name] = handler
}

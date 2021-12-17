package tseries

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource"
)

var registeredTS = make(map[string]TSGenerator)

type TSeriesData struct { //nolint
	Measurement string
	Tags        map[string]string
	Fields      map[string]string
	Value       string
	Timestamp   int64
}

type TSeriesRequest struct { //nolint
	Data     interface{}       `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type TSeriesResponse struct { //nolint
	Data     []byte            `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type TimeSerier interface {
	Init(resource.Metadata) error
	Write(ctx context.Context, req *TSeriesRequest) (*TSeriesResponse, error)
}

type TSGenerator func() TimeSerier

func NewTimeSerier(name string) TimeSerier {
	generator, has := registeredTS[name]
	if has {
		return generator()
	}
	return registeredTS["noop"]()
}

func Register(name string, handler TSGenerator) {
	registeredTS[name] = handler
}

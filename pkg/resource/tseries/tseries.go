package tseries

import (
	"context"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

var registeredTS = make(map[string]Generator)

type TSeriesData struct { // nolint
	Measurement string
	Tags        map[string]string
	Fields      map[string]string
	Value       string
	Timestamp   int64
}

type TSeriesRequest struct { // nolint
	Data     interface{}       `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type TSeriesResponse struct { // nolint
	Data     []byte            `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type TimeSerier interface {
	Write(context.Context, *TSeriesRequest) (*TSeriesResponse, error)
}

type Generator func(map[string]interface{}) (TimeSerier, error)

func NewTimeSerier(metadata resource.Metadata) TimeSerier {
	var err error
	var tsClient TimeSerier
	if generator, has := registeredTS[metadata.Name]; has {
		if tsClient, err = generator(metadata.Properties); nil != err {
			log.Debug("new TSDB instance", zfield.Type(metadata.Name))
			return tsClient
		}
		log.Error("new TSDB instance", zap.Error(err),
			zap.String("name", metadata.Name), zap.Any("properties", metadata.Properties))
	}

	log.Warn("new TSDB.noop instance")
	tsClient, _ = registeredTS["noop"](metadata.Properties)
	return tsClient
}

func Register(name string, handler Generator) {
	registeredTS[name] = handler
}

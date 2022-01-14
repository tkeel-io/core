package statem

import (
	"context"

	"github.com/tkeel-io/core/pkg/resource/tseries"
)

type IStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}

type IPubsub interface{}

type ISearch interface {
	Index(ctx context.Context, data map[string]interface{}) error
}

type TSerier interface {
	Write(ctx context.Context, data []tseries.TSeriesData) error
}

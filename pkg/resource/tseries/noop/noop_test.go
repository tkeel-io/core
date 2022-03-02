package noop

import (
	"context"
	"testing"

	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/tseries"
)

func TestNoop(t *testing.T) {
	n := newNoop()
	n.Init(resource.Metadata{})
	n.Write(context.Background(), &tseries.TSeriesRequest{})
}

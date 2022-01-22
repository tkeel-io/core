package noop

import (
	"context"
	"testing"

	"github.com/tkeel-io/core/pkg/resource/tseries"
)

func TestNoop(t *testing.T) {
	n := &noop{}
	n.Write(context.Background(), &tseries.TSeriesRequest{})
}

package timeseries

import (
	"context"
	"testing"

	"github.com/tkeel-io/core/pkg/resource"
)

func TestNoop(t *testing.T) {
	n := newNoop()
	n.Init(resource.Metadata{})
	n.Write(context.Background(), &WriteRequest{})
}

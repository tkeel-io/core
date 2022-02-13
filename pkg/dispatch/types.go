package dispatch

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util"
)

type Dispatcher interface {
	Dispatch(ctx context.Context, ev cloudevents.Event) error
}

var loopbackQueue = &dao.Queue{
	ID:           util.UUID(),
	Name:         "loopback-core-route-and-republish",
	Type:         "loopback",
	Consumers:    []string{},
	ConsumerType: dao.ConsumerTypeDispatch,
	Description:  "used for core.runtime actor republish event and core.APIs route request.",
}

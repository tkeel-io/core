package dispatch

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
)

type Dispatcher interface {
	DispatchToLog(context.Context, []byte) error
	Dispatch(context.Context, v1.Event) error
}

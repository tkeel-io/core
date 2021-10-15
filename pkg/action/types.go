package action

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

// IAction is an interface for action.
type IAction interface {
	Invoke(ctx context.Context, e *common.TopicEvent) (retry bool, err error)
}

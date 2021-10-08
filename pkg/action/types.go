package action

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

//IAction is a interface for action.
type IAction interface {
	Invoke(ctx context.Context, e *common.TopicEvent) (retry bool, err error)
}

// type TaskFactory interface {
// 	NewTask() IAction
// }

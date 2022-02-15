package holder

import (
	"context"

	"github.com/tkeel-io/core/pkg/types"
)

type Holder interface {
	Wait(ctx context.Context, id string) Response
	OnRespond(*Response)
}

type Response struct {
	ID       string
	Status   types.Status
	ErrCode  string
	Metadata map[string]string
	Data     []byte
}

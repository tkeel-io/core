package runtime

import (
	"context"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
	ants "github.com/panjf2000/ants/v2"
	"github.com/tkeel-io/core/pkg/inbox"
)

type MessageRouter struct{}

type Container struct{}

type EntityManager struct {
	// router route message.
	msgRouter MessageRouter
	// inboxes 用于提供数据的可靠性.
	msgInboxes []inbox.Inbox
	// enContainer store entities.
	enContainer Container
	// coroutinePool coroutine pool.
	coroutinePool *ants.Pool
	// daprClient dapr-sidecar client.
	daprClient dapr.Client

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewEntityManager(ctx context.Context) *EntityManager {
	ctx, cancel := context.WithCancel(ctx)
	return &EntityManager{
		msgRouter:     MessageRouter{},
		msgInboxes:    []inbox.Inbox{},
		enContainer:   Container{},
		daprClient:    nil,
		coroutinePool: nil,
		lock:          sync.RWMutex{},
		ctx:           ctx,
		cancel:        cancel,
	}
}

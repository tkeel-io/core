package entities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/statem"
)

type Entity struct {
	stateMarchine statem.StateMarchiner
}

func newEntity(ctx context.Context, mgr *EntityManager, in *statem.Base) (EntityOp, error) {
	stateM, err := statem.NewState(ctx, mgr, in, nil)
	if nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	}

	return &Entity{stateMarchine: stateM}, nil
}

// GetID return state marchine id.
func (e *Entity) GetID() string {
	return e.stateMarchine.GetID()
}

// GetBase returns state.Base.
func (e *Entity) GetBase() *statem.Base {
	return e.stateMarchine.GetBase()
}

// Setup state marchine setup.
func (e *Entity) Setup() error {
	return errors.Wrap(e.stateMarchine.Setup(), "entity setup failed")
}

// OnMessage recv message from pubsub.
func (e *Entity) OnMessage(msg statem.Message) bool {
	return e.stateMarchine.OnMessage(msg)
}

// InvokeMsg dispose entity message.
func (e *Entity) HandleLoop() {
	e.stateMarchine.HandleLoop()
}

// StateManager returns state manager.
func (e *Entity) GetManager() statem.StateManager {
	return e.stateMarchine.GetManager()
}

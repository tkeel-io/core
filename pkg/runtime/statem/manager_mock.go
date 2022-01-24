package statem

import (
	"context"

	"github.com/tkeel-io/core/pkg/runtime/message"
)

type StateManagerMock struct {
}

func NewStateManagerMock() StateManager {
	return &StateManagerMock{}
}

func (s *StateManagerMock) Start() error              { return nil }
func (s *StateManagerMock) Shutdown() error           { return nil }
func (s *StateManagerMock) Resource() ResourceManager { return nil }
func (s *StateManagerMock) RouteMessage(ctx context.Context, msgCtx message.MessageContext) error {
	return nil
}
func (s *StateManagerMock) HandleMessage(ctx context.Context, msgCtx message.MessageContext) error {
	return nil
}

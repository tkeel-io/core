package statem

import (
	"context"
)

type StateManagerMock struct {
}

func NewStateManagerMock() StateManager {
	return &StateManagerMock{}
}

func (s *StateManagerMock) Start() error                                                  { return nil }
func (s *StateManagerMock) Shutdown() error                                               { return nil }
func (s *StateManagerMock) GetResource() ResourceManager                                  { return nil }
func (s *StateManagerMock) RouteMessage(ctx context.Context, msgCtx MessageContext) error { return nil }
func (s *StateManagerMock) HandleMessage(ctx context.Context, msgCtx MessageContext) error {
	return nil
}

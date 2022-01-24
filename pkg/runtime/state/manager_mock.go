package state

import (
	"context"

	"github.com/tkeel-io/core/pkg/runtime/message"
)

type ManagerMock struct {
}

func NewManagerMock() Manager {
	return &ManagerMock{}
}

func (s *ManagerMock) Start() error              { return nil }
func (s *ManagerMock) Shutdown() error           { return nil }
func (s *ManagerMock) Resource() ResourceManager { return nil }
func (s *ManagerMock) RouteMessage(ctx context.Context, msgCtx message.MessageContext) error {
	return nil
}
func (s *ManagerMock) HandleMessage(ctx context.Context, msgCtx message.MessageContext) error {
	return nil
}

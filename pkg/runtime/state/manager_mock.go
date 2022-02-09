package state

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
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
func (s *ManagerMock) RouteMessage(ctx context.Context, e cloudevents.Event) error {
	return nil
}
func (s *ManagerMock) HandleMessage(ctx context.Context, msgCtx message.Context) error {
	return nil
}

func (s *ManagerMock) SetRepublisher(republisher Republisher) {}

package state

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
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
func (s *ManagerMock) HandleMessage(ctx context.Context, e cloudevents.Event) error {
	return nil
}

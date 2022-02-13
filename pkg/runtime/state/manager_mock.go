package state

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/tkeel-io/core/pkg/types"
)

type ManagerMock struct {
}

func NewManagerMock() types.Manager {
	return &ManagerMock{}
}

func (s *ManagerMock) Start() error                    { return nil }
func (s *ManagerMock) Shutdown() error                 { return nil }
func (s *ManagerMock) Resource() types.ResourceManager { return nil }
func (s *ManagerMock) RouteMessage(ctx context.Context, e cloudevents.Event) error {
	return nil
}

func (s *ManagerMock) SetRepublisher(republisher types.Republisher) {}

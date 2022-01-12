package statem

import (
	"context"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/core/pkg/resource/tseries"
)

type StateManagerMock struct {
}

func NewStateManagerMock() StateManager {
	return &StateManagerMock{}
}

func (s *StateManagerMock) Start() error                                                 { return nil }
func (s *StateManagerMock) SendMsg(msgCtx MessageContext)                                {}
func (s *StateManagerMock) GetDaprClient() dapr.Client                                   { return nil }
func (s *StateManagerMock) HandleMsg(ctx context.Context, msgCtx MessageContext)         {}
func (s *StateManagerMock) EscapedEntities(expression string) []string                   { return nil }
func (s *StateManagerMock) SearchFlush(context.Context, map[string]interface{}) error    { return nil }
func (s *StateManagerMock) TimeSeriesFlush(context.Context, []tseries.TSeriesData) error { return nil }
func (s *StateManagerMock) SetConfigs(context.Context, *Base) error                      { return nil }
func (s *StateManagerMock) AppendConfigs(context.Context, *Base) error                   { return nil }
func (s *StateManagerMock) RemoveConfigs(context.Context, *Base, []string) error         { return nil }
func (s *StateManagerMock) PatchConfigs(context.Context, *Base, []*PatchData) error {
	return nil
}

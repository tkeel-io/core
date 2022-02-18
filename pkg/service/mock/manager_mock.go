package mock

import (
	"context"

	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/manager/holder"
	"github.com/tkeel-io/core/pkg/runtime/state"
)

type APIManagerMock struct {
}

func NewAPIManagerMock() apim.APIManager {
	return &APIManagerMock{}
}

// Start start Entity manager.
func (m *APIManagerMock) Start() error { return nil }

// OnMessage handle message.
func (m *APIManagerMock) OnRespond(ctx context.Context, resp *holder.Response) {
}

// CreateEntity create entity.
func (m *APIManagerMock) CreateEntity(ctx context.Context, en *apim.Base) (*apim.Base, error) {
	return en, nil
}

// UpdateEntity create entity.
func (m *APIManagerMock) UpdateEntity(ctx context.Context, en *apim.Base) (*apim.Base, error) {
	return en, nil
}

// DeleteEntity delete entity.
func (m *APIManagerMock) DeleteEntity(ctx context.Context, en *apim.Base) (err error) {
	return nil
}

// GetProperties returns entity properties.
func (m *APIManagerMock) GetEntity(ctx context.Context, en *apim.Base) (base *apim.Base, err error) {
	return en, nil
}

// SetProperties set entity properties.
func (m *APIManagerMock) UpdateEntityProps(ctx context.Context, en *apim.Base) (base *apim.Base, err error) {
	return en, nil
}

// PatchEntity patch entity properties.
func (m *APIManagerMock) PatchEntityProps(ctx context.Context, en *apim.Base, patchData []state.PatchData) (base *apim.Base, err error) {
	return en, nil
}

func (m *APIManagerMock) GetEntityProps(ctx context.Context, en *apim.Base, propertyKeys []string) (*apim.Base, error) {
	return en, nil
}

// SetConfigs set entity configs.
func (m *APIManagerMock) UpdateEntityConfigs(ctx context.Context, en *apim.Base) (base *apim.Base, err error) {
	return en, nil
}

// PatchConfigs patch entity configs.
func (m *APIManagerMock) PatchEntityConfigs(ctx context.Context, en *apim.Base, patchData []state.PatchData) (base *apim.Base, err error) {
	return en, nil
}

// QueryConfigs returns entity configs.
func (m *APIManagerMock) GetEntityConfigs(ctx context.Context, en *apim.Base, propertyIDs []string) (base *apim.Base, err error) {
	return en, nil
}

// RemoveConfigs remove entity configs.
func (m *APIManagerMock) RemoveEntityConfigs(ctx context.Context, en *apim.Base, propertyIDs []string) (base *apim.Base, err error) {
	return en, nil
}

// AppendMapper append entity mapper.
func (m *APIManagerMock) AppendMapper(ctx context.Context, en *apim.Base) error {
	return nil
}

// RemoveMapper remove entity mapper.
func (m *APIManagerMock) RemoveMapper(ctx context.Context, en *apim.Base) error {
	return nil
}

// CheckSubscription check subscription.
func (m *APIManagerMock) CheckSubscription(ctx context.Context, en *apim.Base) (err error) {
	return nil
}

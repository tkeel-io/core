package mock

import (
	"context"

	v1 "github.com/tkeel-io/core/api/core/v1"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/manager/holder"
	"github.com/tkeel-io/core/pkg/repository/dao"
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
func (m *APIManagerMock) CreateEntity(context.Context, *apim.Base) (*apim.BaseRet, error) {
	return nil, nil
}

// UpdateEntity update entity.
func (m *APIManagerMock) PatchEntity(context.Context, *apim.Base, []*v1.PatchData) (*apim.BaseRet, []byte, error) {
	return nil, nil, nil
}

// DeleteEntity delete entity.
func (m *APIManagerMock) DeleteEntity(context.Context, *apim.Base) error {
	return nil
}

// GetProperties returns entity properties.
func (m *APIManagerMock) GetEntity(context.Context, *apim.Base) (*apim.BaseRet, error) {
	return nil, nil
}

// AppendMapper append entity mapper.
func (m *APIManagerMock) AppendMapper(ctx context.Context, mp *dao.Mapper) error {
	return nil
}

// RemoveMapper remove entity mapper.
func (m *APIManagerMock) RemoveMapper(ctx context.Context, mp *dao.Mapper) error {
	return nil
}

func (m *APIManagerMock) GetMapper(context.Context, *dao.Mapper) (*dao.Mapper, error) {
	return &dao.Mapper{}, nil
}

// ListMapper returns entity mappers.
func (m *APIManagerMock) ListMapper(context.Context, *apim.Base) ([]dao.Mapper, error) {
	return []dao.Mapper{}, nil
}

// CheckSubscription check subscription.
func (m *APIManagerMock) CheckSubscription(ctx context.Context, en *apim.Base) (err error) {
	return nil
}

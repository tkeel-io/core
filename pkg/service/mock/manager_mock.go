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
func (m *APIManagerMock) CreateEntity(_ context.Context, in *apim.Base) (*apim.BaseRet, error) {
	return &apim.BaseRet{
		ID:     in.ID,
		Type:   in.Type,
		Owner:  in.Owner,
		Source: in.Source,
	}, nil
}

// UpdateEntity update entity.
func (m *APIManagerMock) PatchEntity(_ context.Context, in *apim.Base, _ []*v1.PatchData, _ ...apim.Option) (*apim.BaseRet, []byte, error) {
	return &apim.BaseRet{
		ID:     in.ID,
		Type:   in.Type,
		Owner:  in.Owner,
		Source: in.Source,
	}, nil, nil
}

// DeleteEntity delete entity.
func (m *APIManagerMock) DeleteEntity(context.Context, *apim.Base) error {
	return nil
}

// GetProperties returns entity properties.
func (m *APIManagerMock) GetEntity(_ context.Context, in *apim.Base) (*apim.BaseRet, error) {
	return &apim.BaseRet{
		ID:     in.ID,
		Type:   in.Type,
		Owner:  in.Owner,
		Source: in.Source,
	}, nil
}

// AppendMapper append entity mapper.
func (m *APIManagerMock) AppendMapper(ctx context.Context, mp *dao.Mapper) error {
	return nil
}

// AppendMapperZ append entity mapper.
func (m *APIManagerMock) AppendMapperZ(ctx context.Context, mp *dao.Mapper) error {
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

func (m *APIManagerMock) AppendExpression(context.Context, []dao.Expression) error { return nil }
func (m *APIManagerMock) RemoveExpression(context.Context, []dao.Expression) error { return nil }

func (m *APIManagerMock) GetExpression(context.Context, dao.Expression) (*dao.Expression, error) {
	return nil, nil
}

func (m *APIManagerMock) ListExpression(context.Context, *apim.Base) ([]dao.Expression, error) {
	return nil, nil
}

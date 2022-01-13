package mock

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type EntityManagerMock struct {
}

func NewEntityManagerMock() entities.EntityManager {
	return &EntityManagerMock{}
}

// Start start Entity manager.
func (m *EntityManagerMock) Start() error { return nil }

// OnMessage handle message.
func (m *EntityManagerMock) OnMessage(ctx context.Context, msgCtx statem.MessageContext) {
	log.Debug("handle message", zap.Any("headers", msgCtx.Headers), zap.Any("message", msgCtx.Message))
}

// CreateEntity create entity.
func (m *EntityManagerMock) CreateEntity(ctx context.Context, en *statem.Base) (*statem.Base, error) {
	return en, nil
}

// DeleteEntity delete entity.
func (m *EntityManagerMock) DeleteEntity(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return en, nil
}

// GetProperties returns entity properties.
func (m *EntityManagerMock) GetProperties(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return en, nil
}

// SetProperties set entity properties.
func (m *EntityManagerMock) SetProperties(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return en, nil
}

// PatchEntity patch entity properties.
func (m *EntityManagerMock) PatchEntity(ctx context.Context, en *statem.Base, patchData []*pb.PatchData) (base *statem.Base, err error) {
	return en, nil
}

// AppendMapper append entity mapper.
func (m *EntityManagerMock) AppendMapper(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return en, nil
}

// RemoveMapper remove entity mapper.
func (m *EntityManagerMock) RemoveMapper(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return en, nil
}

// CheckSubscription check subscription.
func (m *EntityManagerMock) CheckSubscription(ctx context.Context, en *statem.Base) (err error) {
	return nil
}

// SetConfigs set entity configs.
func (m *EntityManagerMock) SetConfigs(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return en, nil
}

// PatchConfigs patch entity configs.
func (m *EntityManagerMock) PatchConfigs(ctx context.Context, en *statem.Base, patchData []*statem.PatchData) (base *statem.Base, err error) {
	return en, nil
}

// AppendConfigs append entity configs.
func (m *EntityManagerMock) AppendConfigs(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return en, nil
}

// RemoveConfigs remove entity configs.
func (m *EntityManagerMock) RemoveConfigs(ctx context.Context, en *statem.Base, propertyIDs []string) (base *statem.Base, err error) {
	return en, nil
}

// QueryConfigs returns entity configs.
func (m *EntityManagerMock) QueryConfigs(ctx context.Context, en *statem.Base, propertyIDs []string) (base *statem.Base, err error) {
	return en, nil
}

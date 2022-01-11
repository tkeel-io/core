package mock

import (
	"context"
	"time"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/statem"
)

type EntityManagerMock struct {
}

func NewEntityManagerMock() entities.EntityManager {
	return &EntityManagerMock{}
}

// Start start Entity manager.
func (m *EntityManagerMock) Start() error { return nil }

// OnMessage handle message.
func (m *EntityManagerMock) OnMessage(ctx context.Context, msgCtx statem.MessageContext) {}

// CreateEntity create entity.
func (m *EntityManagerMock) CreateEntity(ctx context.Context, base *statem.Base) (*statem.Base, error) {
	return nil, nil
}

// DeleteEntity delete entity.
func (m *EntityManagerMock) DeleteEntity(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return nil, nil
}

// GetProperties returns entity properties.
func (m *EntityManagerMock) GetProperties(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return &statem.Base{
		ID:           "device123",
		Type:         "DEVICE",
		Owner:        "admin",
		Source:       "dm",
		Version:      0,
		LastTime:     time.Now().UnixMilli(),
		Mappers:      []statem.MapperDesc{{Name: "mapper123", TQLString: "insert into device123 select device234.temp as temp"}},
		KValues:      map[string]constraint.Node{"temp": constraint.NewNode(25)},
		ConfigsBytes: nil,
	}, nil
}

// SetProperties set entity properties.
func (m *EntityManagerMock) SetProperties(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return nil, nil
}

// PatchEntity patch entity properties.
func (m *EntityManagerMock) PatchEntity(ctx context.Context, en *statem.Base, patchData []*pb.PatchData) (base *statem.Base, err error) {
	return nil, nil
}

// AppendMapper append entity mapper.
func (m *EntityManagerMock) AppendMapper(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return nil, nil
}

// RemoveMapper remove entity mapper.
func (m *EntityManagerMock) RemoveMapper(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return nil, nil
}

// CheckSubscription check subscription.
func (m *EntityManagerMock) CheckSubscription(ctx context.Context, en *statem.Base) (err error) {
	return nil
}

// SetConfigs set entity configs.
func (m *EntityManagerMock) SetConfigs(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return nil, nil
}

// PatchConfigs patch entity configs.
func (m *EntityManagerMock) PatchConfigs(ctx context.Context, en *statem.Base, patchData []*statem.PatchData) (base *statem.Base, err error) {
	return nil, nil
}

// AppendConfigs append entity configs.
func (m *EntityManagerMock) AppendConfigs(ctx context.Context, en *statem.Base) (base *statem.Base, err error) {
	return nil, nil
}

// RemoveConfigs remove entity configs.
func (m *EntityManagerMock) RemoveConfigs(ctx context.Context, en *statem.Base, propertyIDs []string) (base *statem.Base, err error) {
	return nil, nil
}

// QueryConfigs returns entity configs.
func (m *EntityManagerMock) QueryConfigs(ctx context.Context, en *statem.Base, propertyIDs []string) (base *statem.Base, err error) {
	return nil, nil
}

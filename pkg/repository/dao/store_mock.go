package dao

// import (
// 	"context"

// 	"github.com/pkg/errors"
// 	"github.com/tkeel-io/core/pkg/constraint"
// 	"github.com/tkeel-io/core/pkg/resource/state"
// )

// type storeMock struct {
// 	entityCodec entityCodec
// }

// // GetState retrieves state from specific store using default consistency option.
// func (m *storeMock) Get(ctx context.Context, key string) (item *state.StateItem, err error) {
// 	bytes, err := m.entityCodec.Encode(&Entity{
// 		ID:         "device123",
// 		Type:       "DEVICE",
// 		Owner:      "admin",
// 		Source:     "dm",
// 		Version:    0,
// 		Properties: map[string]constraint.Node{"temp": constraint.NewNode(25)},
// 	})

// 	return &state.StateItem{
// 		Key:   key,
// 		Value: bytes,
// 	}, errors.Wrap(err, "store mock")
// }

// // SaveState saves the raw data into store using default state options.
// func (m *storeMock) Set(ctx context.Context, key string, data []byte) error {
// 	return nil
// }

// // Del delete record from store.
// func (m *storeMock) Del(ctx context.Context, key string) error {
// 	return nil
// }

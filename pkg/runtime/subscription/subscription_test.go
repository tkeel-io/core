package subscription

import (
	"testing"
)

func TestSubscription(t *testing.T) {
	// sub, err := newSubscription(context.Background(), nil, &statem.Base{
	// 	ID:     "sub123",
	// 	Type:   "SUBCRIPTION",
	// 	Owner:  "admin",
	// 	Source: "device-manager",
	// 	KValues: map[string]constraint.Node{
	// 		"mode":        constraint.NewNode("realtime"),
	// 		"source":      constraint.NewNode("device-manager"),
	// 		"filter":      constraint.NewNode("insert into sub123 select device123.temp"),
	// 		"target":      constraint.NewNode("device123"),
	// 		"topic":       constraint.NewNode("core-sub123"),
	// 		"pubsub_name": constraint.NewNode("core-pubsub"),
	// 	},
	// })

	// assert.Equal(t, nil, err)
	// assert.Equal(t, "sub123", sub.GetID())
}

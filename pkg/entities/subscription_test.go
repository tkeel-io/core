package entities

import (
	"context"
	"testing"

	ants "github.com/panjf2000/ants/v2"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/statem"
)

func TestSubscriptionCreate(t *testing.T) {
	coroutinePool, err := ants.NewPool(500)
	if nil != err {
		panic(err)
	}

	mgr, _ := NewEntityManager(context.Background(), coroutinePool)

	en := &statem.Base{
		ID:     "",
		Type:   EntityTypeSubscription,
		Owner:  "tomas",
		Source: "PluginB",
		KValues: map[string]constraint.Node{
			SubscriptionFieldMode:   constraint.RawNode(SubscriptionModeRealtime),
			SubscriptionFieldSource: constraint.RawNode("PluginA"),
			SubscriptionFieldTarget: constraint.RawNode("PluginA"),
			SubscriptionFieldFilter: constraint.RawNode("select *"),
		},
	}

	sub, err := newSubscription(context.Background(), mgr, en)

	t.Log("mapstructure: ", sub, err)
	t.Log("subscription status: ")
}

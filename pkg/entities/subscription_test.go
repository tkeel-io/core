package entities

import (
	"context"
	"testing"

	ants "github.com/panjf2000/ants/v2"
)

func TestSubscriptionCreate(t *testing.T) {
	coroutinePool, err := ants.NewPool(500)
	if nil != err {
		panic(err)
	}

	mgr, _ := NewEntityManager(context.Background(), coroutinePool)

	en := &EntityBase{
		ID:       "",
		Type:     EntityTypeSubscription,
		Owner:    "tomas",
		PluginID: "PluginB",
		KValues: map[string]interface{}{
			SubscriptionFieldMode:   SubscriptionModeRealtime,
			SubscriptionFieldSource: "PluginA",
			SubscriptionFieldTarget: "PluginA",
			SubscriptionFieldFilter: "select *",
		},
	}

	sub, err := newSubscription(context.Background(), mgr, en)

	t.Log("mapstructure: ", sub.SubscriptionBase, err)
	t.Log("subscription status: ", sub.Status)
}

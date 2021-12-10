/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

	mgr, _ := NewEntityManager(context.Background(), coroutinePool, nil)

	en := &statem.Base{
		ID:     "",
		Type:   EntityTypeSubscription,
		Owner:  "tomas",
		Source: "PluginB",
		KValues: map[string]constraint.Node{
			SubscriptionFieldMode:   constraint.NewNode(SubscriptionModeRealtime),
			SubscriptionFieldSource: constraint.NewNode("PluginA"),
			SubscriptionFieldTarget: constraint.NewNode("PluginA"),
			SubscriptionFieldFilter: constraint.NewNode("select *"),
		},
	}

	sub, err := newSubscription(context.Background(), mgr, en)

	t.Log("mapstructure: ", sub, err)
	t.Log("subscription status: ")
}

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

package subscription

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const (
	// subscription mode enum.
	SubscriptionModeRealtime = "REALTIME"
	SubscriptionModePeriod   = "PERIOD"
	SubscriptionModeChanged  = "ONCHANGED"
	SubscriptionModeUndefine = "UNDEFINE"

	// subscription required fileds.
	SubscriptionFieldMode       = "mode"
	SubscriptionFieldFilter     = "filter"
	SubscriptionFieldTopic      = "topic"
	SubscriptionFieldPubsubName = "pubsub_name"
)

// subscription subscription actor based entity.
type subscription struct {
	pubsubClient pubsub.Pubsub
	stateMachine statem.StateMachiner
	stateManager statem.StateManager
}

// NewSubscription returns a subscription.
func NewSubscription(ctx context.Context, mgr statem.StateManager, in *dao.Entity) (stateM statem.StateMachiner, err error) {
	subsc := subscription{stateManager: mgr}
	errFunc := func(err error) error { return errors.Wrap(err, "create subscription") }

	if stateM, err = statem.NewState(ctx, mgr, in, subsc.HandleMessage); nil != err {
		return nil, errFunc(err)
	}

	// decode in.Properties into subsc.
	subsc.stateMachine = stateM
	if err = subsc.checkSubscription(); nil != err {
		return nil, errFunc(err)
	}

	// create pubsub client.
	subsc.pubsubClient = pubsub.NewPubsub(resource.Metadata{
		Name: "dapr",
		Properties: map[string]interface{}{
			"pubsub_name": subsc.PubsubName(),
			"topic_name":  subsc.Topic(),
		},
	})

	return &subsc, nil
}

func (s *subscription) Flush(ctx context.Context) error {
	return errors.Wrap(s.stateMachine.Flush(ctx), "flush subscription")
}

// GetID return state machine id.
func (s *subscription) GetID() string {
	return s.stateMachine.GetID()
}

// GetMode returns subscription mode.
func (s *subscription) GetMode() string {
	return s.Mode()
}

func (s *subscription) GetStatus() statem.Status {
	return s.stateMachine.GetStatus()
}

func (s *subscription) GetEntity() *dao.Entity {
	return s.stateMachine.GetEntity()
}

func (s *subscription) WithContext(sCtx statem.StateContext) statem.StateMachiner {
	return s.stateMachine.WithContext(sCtx)
}

// OnMessage recv message from pubsub.
func (s *subscription) OnMessage(msg message.Message) bool {
	return s.stateMachine.OnMessage(msg)
}

// InvokeMsg dispose entity message.
func (s *subscription) HandleLoop() {
	s.stateMachine.HandleLoop()
}

func (s *subscription) HandleMessage(m message.Message) []mapper.WatchKey {
	log.Debug("on subscribe", zap.String("subscription", s.GetID()), logger.Message(m))
	var watchKeys []mapper.WatchKey
	switch msg := m.(type) {
	case message.PropertyMessage:
		switch s.Mode() {
		case SubscriptionModeRealtime:
			watchKeys = s.invokeRealtime(msg)
		case SubscriptionModePeriod:
			watchKeys = s.invokePeriod(msg)
		case SubscriptionModeChanged:
			watchKeys = s.invokeChanged(msg)
		default:
			// invalid subscription mode.
			log.Error("undefine subscription mode, mode.", zap.String("mode", s.Mode()))
		}
	default:
		// invalid msg typs.
		log.Error("undefine message type.", logger.Message(msg))
	}
	return watchKeys
}

// invokeRealtime invoke property where mode is realtime.
func (s *subscription) invokeRealtime(msg message.PropertyMessage) []mapper.WatchKey {
	b := s.stateMachine.GetEntity()
	cp := dao.Entity{
		ID:         b.ID,
		Type:       b.Type,
		Owner:      b.Owner,
		Source:     b.Source,
		Version:    b.Version,
		LastTime:   b.LastTime,
		TemplateID: b.TemplateID,
		Properties: make(map[string]constraint.Node),
	}

	cp.Properties = msg.Properties
	if err := s.pubsubClient.Send(context.Background(), cp); nil != err {
		log.Error("invoke realtime subscription failed.", logger.Message(msg), zap.Error(err))
	}

	return nil
}

// invokePeriod.
func (s *subscription) invokePeriod(msg message.PropertyMessage) []mapper.WatchKey {
	b := s.stateMachine.GetEntity()
	cp := dao.Entity{
		ID:         b.ID,
		Type:       b.Type,
		Owner:      b.Owner,
		Source:     b.Source,
		Version:    b.Version,
		LastTime:   b.LastTime,
		TemplateID: b.TemplateID,
		Properties: make(map[string]constraint.Node),
	}

	cp.Properties = msg.Properties
	if err := s.pubsubClient.Send(context.Background(), cp); nil != err {
		log.Error("invoke period subscription failed.", logger.Message(msg), zap.Error(err))
	}

	return nil
}

// invokeChanged.
func (s *subscription) invokeChanged(msg message.PropertyMessage) []mapper.WatchKey {
	b := s.stateMachine.GetEntity()
	cp := dao.Entity{
		ID:         b.ID,
		Type:       b.Type,
		Owner:      b.Owner,
		Source:     b.Source,
		Version:    b.Version,
		LastTime:   b.LastTime,
		TemplateID: b.TemplateID,
		Properties: make(map[string]constraint.Node),
	}

	cp.Properties = msg.Properties
	if err := s.pubsubClient.Send(context.Background(), cp); nil != err {
		log.Error("invoke changed subscription failed.", logger.Message(msg), zap.Error(err))
	}

	return nil
}

func (s *subscription) checkSubscription() error {
	sb := Base{}
	decode2Subscription(s.stateMachine.GetEntity().Properties, &sb)
	return errors.Wrap(sb.Validate(), "check subscription required fileds")
}

func (s *subscription) Mode() string {
	sb := Base{}
	decode2Subscription(s.stateMachine.GetEntity().Properties, &sb)
	return sb.Mode
}

func (s *subscription) Filter() string {
	sb := Base{}
	decode2Subscription(s.stateMachine.GetEntity().Properties, &sb)
	return sb.Filter
}

func (s *subscription) Topic() string {
	sb := Base{}
	decode2Subscription(s.stateMachine.GetEntity().Properties, &sb)
	return sb.Topic
}

func (s *subscription) PubsubName() string {
	sb := Base{}
	decode2Subscription(s.stateMachine.GetEntity().Properties, &sb)
	return sb.PubsubName
}

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
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/runtime/state"
	"github.com/tkeel-io/core/pkg/util"
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
	stateMachine state.Machiner
	stateManager state.Manager
}

// NewSubscription returns a subscription.
func NewSubscription(ctx context.Context, mgr state.Manager, in *dao.Entity) (stateM state.Machiner, err error) {
	subsc := subscription{stateManager: mgr}
	errFunc := func(err error) error { return errors.Wrap(err, "create subscription") }

	if stateM, err = state.NewState(ctx, mgr, in, subsc.HandleMessage); nil != err {
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

func (s *subscription) GetStatus() state.Status {
	return s.stateMachine.GetStatus()
}

func (s *subscription) GetEntity() *dao.Entity {
	return s.stateMachine.GetEntity()
}

func (s *subscription) WithContext(sCtx state.StateContext) state.Machiner {
	return s.stateMachine.WithContext(sCtx)
}

// OnMessage recv message from pubsub.
func (s *subscription) OnMessage(msgCtx message.Context) bool {
	return s.stateMachine.OnMessage(msgCtx)
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
	reqID := util.UUID()
	msgID := util.UUID()
	eventID := util.UUID()
	sender := fmt.Sprintf("Core.Subscription.%s", s.GetID())

	ev := cloudevents.NewEvent()
	ev.SetID(eventID)
	ev.SetType("core.APIs")
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtRequestID, reqID)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtMessageSender, sender)
	ev.SetExtension(message.ExtEntityID, msg.StateID)
	ev.SetExtension(message.ExtMessageReceiver, s.PubsubName())

	// set data.
	ev.SetDataContentType(cloudevents.Binary)

	bytes, err := constraint.EncodeJSON(msg.Properties)
	if nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("encode properties", logger.Message(msg), zap.Error(err))
		return nil
	}
	ev.SetData(bytes)

	if err := s.pubsubClient.Send(context.Background(), ev); nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("invoke realtime subscription", logger.Message(msg), zap.Error(err))
	}

	return nil
}

// invokePeriod.
func (s *subscription) invokePeriod(msg message.PropertyMessage) []mapper.WatchKey {
	reqID := util.UUID()
	msgID := util.UUID()
	eventID := util.UUID()
	sender := fmt.Sprintf("Core.Subscription.%s", s.GetID())

	ev := cloudevents.NewEvent()
	ev.SetID(eventID)
	ev.SetType("core.APIs")
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtRequestID, reqID)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtMessageSender, sender)
	ev.SetExtension(message.ExtEntityID, msg.StateID)
	ev.SetExtension(message.ExtMessageReceiver, s.PubsubName())

	// set data.
	ev.SetDataContentType(cloudevents.Binary)

	bytes, err := constraint.EncodeJSON(msg.Properties)
	if nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("encode properties", logger.Message(msg), zap.Error(err))
		return nil
	}
	ev.SetData(bytes)

	if err := s.pubsubClient.Send(context.Background(), ev); nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("invoke realtime subscription", logger.Message(msg), zap.Error(err))
	}

	return nil
}

// invokeChanged.
func (s *subscription) invokeChanged(msg message.PropertyMessage) []mapper.WatchKey {
	reqID := util.UUID()
	msgID := util.UUID()
	eventID := util.UUID()
	sender := fmt.Sprintf("Core.Subscription.%s", s.GetID())

	ev := cloudevents.NewEvent()
	ev.SetID(eventID)
	ev.SetType("core.APIs")
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtRequestID, reqID)
	ev.SetExtension(message.ExtMessageID, msgID)
	ev.SetExtension(message.ExtMessageSender, sender)
	ev.SetExtension(message.ExtEntityID, msg.StateID)
	ev.SetExtension(message.ExtMessageReceiver, s.PubsubName())

	// set data.
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	bytes, err := constraint.EncodeJSON(msg.Properties)
	if nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("encode properties", logger.Message(msg), zap.Error(err))
		return nil
	}
	ev.SetData(bytes)

	if err := s.pubsubClient.Send(context.Background(), ev); nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("invoke realtime subscription", logger.Message(msg), zap.Error(err))
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

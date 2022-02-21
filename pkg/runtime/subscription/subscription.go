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
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
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
	pubsubClient     pubsub.Pubsub
	stateMachine     state.Machiner
	republishHandler state.MessageHandler
}

// NewSubscription returns a subscription.
func NewSubscription(ctx context.Context, in *dao.Entity) (stateM state.Machiner, err error) {
	subsc := subscription{}
	errFunc := func(err error) error { return errors.Wrap(err, "create subscription") }

	if stateM, err = state.NewState(ctx, in, nil, nil, subsc.HandleMessage); nil != err {
		return nil, errFunc(err)
	}

	// decode in.Properties into subsc.
	subsc.stateMachine = stateM
	if err = subsc.checkSubscription(); nil != err {
		return nil, errFunc(err)
	}

	// bind republish handler.
	switch subsc.Mode() {
	case SubscriptionModeRealtime:
		subsc.republishHandler = subsc.invokeRealtime
	case SubscriptionModePeriod:
		subsc.republishHandler = subsc.invokePeriod
	case SubscriptionModeChanged:
		subsc.republishHandler = subsc.invokeChanged
	default:
		// invalid subscription mode.
		log.Error("undefine subscription mode, mode.",
			zap.String("mode", subsc.Mode()), zap.Any("entity", in))
		return nil, errFunc(xerrors.ErrInvalidSubscriptionMode)
	}

	// create pubsub client.
	subsc.pubsubClient = pubsub.NewPubsub(
		util.UUID(),
		resource.Metadata{
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

func (s *subscription) Context() *state.StateContext {
	return s.stateMachine.Context()
}

// Invoke message from pubsub.
func (s *subscription) Invoke(ctx context.Context, msgCtx message.Context) error {
	return errors.Wrap(s.stateMachine.Invoke(ctx, msgCtx), "subscription invoke message")
}

func (s *subscription) HandleMessage(msgCtx message.Context) []mapper.WatchKey {
	log.Debug("on subscribe", zfield.ID(s.GetID()), zfield.Message(msgCtx))
	return s.republishHandler(msgCtx)
}

const eventType = "Core.Subscription"

func eventSender(sender string) string {
	return fmt.Sprintf("%s.%s", eventType, sender)
}

// invokeRealtime invoke property where mode is realtime.
func (s *subscription) invokeRealtime(msgCtx message.Context) []mapper.WatchKey {
	var (
		eventID = util.UUID()
		entity  = s.GetEntity()
		ev      = cloudevents.NewEvent()
	)

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtEntityID, entity.ID)
	ev.SetExtension(message.ExtEntityType, entity.Type)
	ev.SetExtension(message.ExtEntityOwner, entity.Owner)
	ev.SetExtension(message.ExtEntitySource, entity.Source)
	ev.SetExtension(message.ExtTemplateID, entity.TemplateID)
	ev.SetExtension(message.ExtMessageReceiver, s.PubsubName())
	ev.SetExtension(message.ExtMessageSender, eventSender(s.GetID()))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	ev.SetData(msgCtx.Message())
	if err := s.pubsubClient.Send(context.Background(), ev); nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("invoke realtime subscription",
			zfield.Message(msgCtx), zap.Error(err))
	}

	return nil
}

// invokePeriod.
func (s *subscription) invokePeriod(msgCtx message.Context) []mapper.WatchKey {
	var (
		eventID = util.UUID()
		entity  = s.GetEntity()
		ev      = cloudevents.NewEvent()
	)

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtEntityID, entity.ID)
	ev.SetDataContentType(cloudevents.ApplicationJSON)
	ev.SetExtension(message.ExtEntityType, entity.Type)
	ev.SetExtension(message.ExtEntityOwner, entity.Owner)
	ev.SetExtension(message.ExtEntitySource, entity.Source)
	ev.SetExtension(message.ExtTemplateID, entity.TemplateID)
	ev.SetExtension(message.ExtMessageReceiver, s.PubsubName())
	ev.SetExtension(message.ExtMessageSender, eventSender(s.GetID()))

	ev.SetData(msgCtx.Message())
	if err := s.pubsubClient.Send(context.Background(), ev); nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("invoke period subscription",
			zfield.Message(msgCtx), zap.Error(err))
	}

	return nil
}

// invokeChanged.
func (s *subscription) invokeChanged(msgCtx message.Context) []mapper.WatchKey {
	var (
		eventID = util.UUID()
		entity  = s.GetEntity()
		ev      = cloudevents.NewEvent()
	)

	ev.SetID(eventID)
	ev.SetType(eventType)
	ev.SetSource(config.Get().Server.Name)
	ev.SetExtension(message.ExtEntityID, entity.ID)
	ev.SetExtension(message.ExtEntityType, entity.Type)
	ev.SetExtension(message.ExtEntityOwner, entity.Owner)
	ev.SetExtension(message.ExtEntitySource, entity.Source)
	ev.SetExtension(message.ExtTemplateID, entity.TemplateID)
	ev.SetExtension(message.ExtMessageReceiver, s.PubsubName())
	ev.SetExtension(message.ExtMessageSender, eventSender(s.GetID()))
	ev.SetDataContentType(cloudevents.ApplicationJSON)

	ev.SetData(msgCtx.Message())
	if err := s.pubsubClient.Send(context.Background(), ev); nil != err {
		// TODO: 对于发送失败的消息需要重新处理.
		log.Error("invoke onchanged subscription",
			zfield.Message(msgCtx), zap.Error(err))
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

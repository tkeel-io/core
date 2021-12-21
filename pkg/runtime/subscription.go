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

package runtime

import (
	"context"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const (
	// subscription mode enum.
	SubscriptionModeUndefine = "undefine"
	SubscriptionModeRealtime = "realtime"
	SubscriptionModePeriod   = "period"
	SubscriptionModeChanged  = "changed"

	// subscription required fileds.
	SubscriptionFieldMode       = "mode"
	SubscriptionFieldSource     = "source"
	SubscriptionFieldTarget     = "target"
	SubscriptionFieldFilter     = "filter"
	SubscriptionFieldTopic      = "topic"
	SubscriptionFieldPubsubName = "pubsub_name"
)

// SubscriptionBase subscription basic information.
type SubscriptionBase struct {
	Mode       string `json:"mode" mapstructure:"mode"`
	Source     string `json:"source" mapstructure:"source"`
	Filter     string `json:"filter" mapstructure:"filter"`
	Target     string `json:"target" mapstructure:"target"`
	Topic      string `json:"topic" mapstructure:"topic"`
	PubsubName string `json:"pubsub_name" mapstructure:"pubsub_name"`
}

// subscription subscription actor based entity.
type subscription struct {
	SubscriptionBase `mapstructure:",squash"`
	daprClient       dapr.Client
	stateMarchine    statem.StateMarchiner `mapstructure:"-"`
}

// newSubscription returns a subscription.
func newSubscription(ctx context.Context, mgr *Manager, in *statem.Base) (statem.StateMarchiner, error) {
	subsc := subscription{
		SubscriptionBase: SubscriptionBase{
			Mode: SubscriptionModeUndefine,
		},
	}

	stateM, err := statem.NewState(ctx, mgr, in, subsc.HandleMessage)
	if nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	} else if err = mapstructure.Decode(in.KValues, &subsc); nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	} else if err = subsc.checkSubscription(); nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	}

	daprClient, err := dapr.NewClient()
	if nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	} else if err = subsc.checkSubscription(); nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	}

	subsc.daprClient = daprClient
	subsc.stateMarchine = stateM
	subsc.GetBase().KValues = in.KValues
	return &subsc, nil
}

func (s *subscription) Flush(ctx context.Context) error {
	return errors.Wrap(s.stateMarchine.Flush(ctx), "flush subscription")
}

// Setup setup filter.
func (s *subscription) Setup() error {
	// set mapper.
	s.stateMarchine.GetBase().Mappers =
		[]statem.MapperDesc{
			{
				Name:      "subscription",
				TQLString: s.Filter,
			},
		}

	return errors.Wrap(s.stateMarchine.Setup(), "subscription setup failed")
}

// GetID return state marchine id.
func (s *subscription) GetID() string {
	return s.stateMarchine.GetID()
}

// GetMode returns subscription mode.
func (s *subscription) GetMode() string {
	return s.Mode
}

func (s *subscription) GetBase() *statem.Base {
	return s.stateMarchine.GetBase()
}

func (s *subscription) SetStatus(status statem.Status) {
	s.stateMarchine.SetStatus(status)
}

func (s *subscription) GetStatus() statem.Status {
	return s.stateMarchine.GetStatus()
}

func (s *subscription) GetManager() statem.StateManager {
	return s.stateMarchine.GetManager()
}

func (s *subscription) SetConfig(configs map[string]constraint.Config) error {
	return errors.Wrap(s.stateMarchine.SetConfig(configs), "subscription.SetConfig failed")
}

// OnMessage recv message from pubsub.
func (s *subscription) OnMessage(msg statem.Message) bool {
	return s.stateMarchine.OnMessage(msg)
}

// InvokeMsg dispose entity message.
func (s *subscription) HandleLoop() {
	s.stateMarchine.HandleLoop()
}

func (s *subscription) HandleMessage(message statem.Message) []WatchKey {
	log.Debug("on subscribe", zap.String("subscription", s.GetID()), logger.MessageInst(message))
	var watchKeys []WatchKey
	switch msg := message.(type) {
	case statem.PropertyMessage:
		switch s.Mode {
		case SubscriptionModeRealtime:
			watchKeys = s.invokeRealtime(msg)
		case SubscriptionModePeriod:
			watchKeys = s.invokePeriod(msg)
		case SubscriptionModeChanged:
			watchKeys = s.invokeChanged(msg)
		default:
			// invalid subscription mode.
			log.Error("undefine subscription mode, mode.", zap.String("mode", s.Mode))
		}
	default:
		// invalid msg typs.
		log.Error("undefine message type.", logger.MessageInst(msg))
	}

	return watchKeys
}

// invokeRealtime invoke property where mode is realtime.
func (s *subscription) invokeRealtime(msg statem.PropertyMessage) []WatchKey {
	// 对于 Realtime 直接转发就OK了.
	base := s.GetBase()
	base.KValues = msg.Properties
	if bytes, err := statem.EncodeBase(base); nil != err {
		log.Error("invoke realtime subscription failed.", logger.MessageInst(msg), zap.Error(err))
	} else if err = s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes); nil != err {
		log.Error("invoke realtime subscription failed.", logger.MessageInst(msg), zap.Error(err))
	}

	return nil
}

// invokePeriod.
func (s *subscription) invokePeriod(msg statem.PropertyMessage) []WatchKey {
	// 对于 Period 直接查询快照.
	base := s.GetBase()
	base.KValues = msg.Properties
	if bytes, err := statem.EncodeBase(base); nil != err {
		log.Error("invoke period subscription failed.", logger.MessageInst(msg), zap.Error(err))
	} else if err = s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes); nil != err {
		log.Error("invoke period subscription failed.", logger.MessageInst(msg), zap.Error(err))
	}

	return nil
}

// invokeChanged.
func (s *subscription) invokeChanged(msg statem.PropertyMessage) []WatchKey {
	// 对于 Changed 直接转发就OK了.
	base := s.GetBase()
	base.KValues = msg.Properties
	if bytes, err := statem.EncodeBase(base); nil != err {
		log.Error("invoke changed subscription failed.", logger.MessageInst(msg), zap.Error(err))
	} else if err = s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes); nil != err {
		log.Error("invoke changed subscription failed.", logger.MessageInst(msg), zap.Error(err))
	}

	return nil
}

// checkSubscription returns subscription status.
func (s *subscription) checkSubscription() error {
	if s.Mode == SubscriptionModeUndefine || s.Source == "" ||
		s.Filter == "" || s.Topic == "" || s.PubsubName == "" {
		return ErrSubscriptionInvalid
	}

	return nil
}

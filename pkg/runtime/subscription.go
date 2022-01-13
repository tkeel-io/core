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
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/environment"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

const (
	// subscription mode enum.
	SubscriptionModeRealtime = "realtime"
	SubscriptionModePeriod   = "period"
	SubscriptionModeChanged  = "changed"
	SubscriptionModeUndefine = "undefine"
)

// SubscriptionBase subscription basic information.
type SubscriptionBase struct {
	Mode       string `json:"mode" mapstructure:"mode"`
	Source     string `json:"source" mapstructure:"source"`
	Filter     string `json:"filter" mapstructure:"filter"`
	Topic      string `json:"topic" mapstructure:"topic"`
	PubsubName string `json:"pubsub_name" mapstructure:"pubsub_name"`
}

// subscription subscription actor based entity.
type subscription struct {
	SubscriptionBase `mapstructure:",squash"`
	daprClient       dapr.Client
	stateMachine     statem.StateMachiner `mapstructure:"-"`
}

func decode2Subscription(kvalues map[string]constraint.Node, subsc *SubscriptionBase) {
	// parse Mode.
	if node, has := kvalues[SubscriptionFieldMode]; has {
		subsc.Mode = node.String()
	}
	// parse Filter.
	if node, has := kvalues[SubscriptionFieldFilter]; has {
		subsc.Filter = node.String()
	}
	// parse Topic.
	if node, has := kvalues[SubscriptionFieldTopic]; has {
		subsc.Topic = node.String()
	}
	// parse PubsubName.
	if node, has := kvalues[SubscriptionFieldPubsubName]; has {
		subsc.PubsubName = node.String()
	}
}

// newSubscription returns a subscription.
func newSubscription(ctx context.Context, mgr *Manager, in *statem.Base) (stateM statem.StateMachiner, err error) {
	subsc := subscription{SubscriptionBase: SubscriptionBase{
		Mode: SubscriptionModeUndefine,
	}}

	errFunc := func(err error) error { return errors.Wrap(err, "create subscription failed") }
	if stateM, err = statem.NewState(ctx, mgr, in, subsc.HandleMessage); nil != err {
		return nil, errFunc(err)
	}

	// decode in.KValues into subsc.
	decode2Subscription(in.KValues, &subsc.SubscriptionBase)
	if err = subsc.checkSubscription(); nil != err {
		return nil, errFunc(err)
	}

	var daprClient dapr.Client
	if daprClient, err = dapr.NewClient(); nil != err {
		return nil, errFunc(err)
	}

	subsc.daprClient = daprClient
	subsc.stateMachine = stateM
	subsc.GetBase().KValues = in.KValues

	// set mapper.
	subsc.stateMachine.GetBase().Mappers = []statem.MapperDesc{{
		Name:      "subscription",
		TQLString: subsc.Filter,
	}}
	return &subsc, nil
}

func (s *subscription) Flush(ctx context.Context) error {
	return errors.Wrap(s.stateMachine.Flush(ctx), "flush subscription")
}

// Setup setup filter.
func (s *subscription) Setup() error {
	return errors.Wrap(s.stateMachine.Setup(), "subscription setup")
}

// GetID return state machine id.
func (s *subscription) GetID() string {
	return s.stateMachine.GetID()
}

// GetMode returns subscription mode.
func (s *subscription) GetMode() string {
	return s.Mode
}

func (s *subscription) GetBase() *statem.Base {
	return s.stateMachine.GetBase()
}

func (s *subscription) SetStatus(status statem.Status) {
	s.stateMachine.SetStatus(status)
}

func (s *subscription) GetStatus() statem.Status {
	return s.stateMachine.GetStatus()
}

func (s *subscription) LoadEnvironments(env environment.ActorEnv) {
	s.stateMachine.LoadEnvironments(env)
}

func (s *subscription) GetManager() statem.StateManager {
	return s.stateMachine.GetManager()
}

// SetConfig set entity configs.
func (s *subscription) SetConfigs(configs map[string]constraint.Config) error {
	err := s.stateMachine.SetConfigs(configs)
	return errors.Wrap(err, "set subscription configs")
}

// PatchConfigs set entity configs.
func (s *subscription) PatchConfigs(patchData []*statem.PatchData) error {
	err := s.stateMachine.PatchConfigs(patchData)
	return errors.Wrap(err, "patch subscription configs")
}

// AppendConfig append entity property config.
func (s *subscription) AppendConfigs(configs map[string]constraint.Config) error {
	err := s.stateMachine.AppendConfigs(configs)
	return errors.Wrap(err, "append subscription configs")
}

// RemoveConfig remove entity property configs.
func (s *subscription) RemoveConfigs(propertyIDs []string) error {
	err := s.stateMachine.RemoveConfigs(propertyIDs)
	return errors.Wrap(err, "remove subscription configs")
}

// OnMessage recv message from pubsub.
func (s *subscription) OnMessage(msg statem.Message) bool {
	return s.stateMachine.OnMessage(msg)
}

// InvokeMsg dispose entity message.
func (s *subscription) HandleLoop() {
	s.stateMachine.HandleLoop()
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
	base := s.GetBase().DuplicateExpectValue()
	base.KValues = msg.Properties
	if err := s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, base); nil != err {
		log.Error("invoke realtime subscription failed.", logger.MessageInst(msg), zap.Error(err))
	}

	return nil
}

// invokePeriod.
func (s *subscription) invokePeriod(msg statem.PropertyMessage) []WatchKey {
	// 对于 Period 直接查询快照.
	base := s.GetBase().DuplicateExpectValue()
	base.KValues = msg.Properties
	if err := s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, base); nil != err {
		log.Error("invoke period subscription failed.", logger.MessageInst(msg), zap.Error(err))
	}

	return nil
}

// invokeChanged.
func (s *subscription) invokeChanged(msg statem.PropertyMessage) []WatchKey {
	// 对于 Changed 直接转发就OK了.
	base := s.GetBase().DuplicateExpectValue()
	base.KValues = msg.Properties
	if err := s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, base); nil != err {
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

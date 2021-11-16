package entities

import (
	"context"
	"encoding/json"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/statem"
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
func newSubscription(ctx context.Context, mgr *EntityManager, in *statem.Base) (statem.StateMarchiner, error) {
	subsc := subscription{
		SubscriptionBase: SubscriptionBase{
			Mode: SubscriptionModeUndefine,
		},
	}

	stateM, err := statem.NewState(ctx, mgr, in, subsc.HandleMessage)
	if nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	}

	stateM.GetBase().Status = subsc.checkSubscription()
	if err = mapstructure.Decode(in.KValues, &subsc); nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	}

	subsc.stateMarchine = stateM
	return &subsc, errors.Wrap(err, "create subscription failed")
}

// Setup setup filter.
func (s *subscription) Setup() error {
	if statem.StateStatusInactive == s.stateMarchine.GetBase().Status {
		return errors.Wrap(errEntityNotAready, "setup subscription failed")
	}

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

func (s *subscription) GetManager() statem.StateManager {
	return s.stateMarchine.GetManager()
}

func (s *subscription) SetConfig(configs map[string]constraint.Config) error {
	return s.stateMarchine.SetConfig(configs)
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
			log.Errorf("undefine subscription mode, mode: %s", s.Mode)
		}
	default:
		// invalid msg typs.
		log.Errorf("undefine message type, msg: %s", msg)
	}

	return watchKeys
}

// invokeRealtime invoke property where mode is realtime.
func (s *subscription) invokeRealtime(msg statem.PropertyMessage) []WatchKey {
	// 对于 Realtime 直接转发就OK了.
	bytes, _ := json.Marshal(msg.Properties)
	if err := s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes); nil != err {
		log.Errorf("invoke realtime subscription failed, msg: %v, %s", msg, err.Error())
	}

	return nil
}

// invokePeriod.
func (s *subscription) invokePeriod(msg statem.PropertyMessage) []WatchKey {
	// 对于 Period 直接查询快照.
	bytes, _ := json.Marshal(msg.Properties)
	if err := s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes); nil != err {
		log.Errorf("invoke realtime subscription failed, msg: %v, %s", msg, err.Error())
	}

	return nil
}

// invokeChanged.
func (s *subscription) invokeChanged(msg statem.PropertyMessage) []WatchKey {
	// 对于 Changed 直接转发就OK了.
	bytes, _ := json.Marshal(msg.Properties)
	if err := s.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes); nil != err {
		log.Errorf("invoke realtime subscription failed, msg: %v, %s", msg, err.Error())
	}

	return nil
}

// checkSubscription returns subscription status.
func (s *subscription) checkSubscription() string {
	if s.Mode == SubscriptionModeUndefine || s.Source == "" ||
		s.Target == "" || s.Filter == "" || s.Topic == "" || s.PubsubName == "" {
		return statem.StateStatusInactive
	}

	return statem.StateStatusActive
}

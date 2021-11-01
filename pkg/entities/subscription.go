package entities

import (
	"context"
	"encoding/json"
	"sync/atomic"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/mapper"
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
	*entity
	SubscriptionBase `mapstructure:",squash"`
}

// newSubscription returns a subscription.
func newSubscription(ctx context.Context, mgr *EntityManager, in *EntityBase) (*subscription, error) {
	en, err := newEntity(ctx, mgr, in)
	if nil != err {
		return nil, errors.Wrap(err, "create subscription failed")
	}

	subsc := subscription{
		entity: en,
		SubscriptionBase: SubscriptionBase{
			Mode: SubscriptionModeUndefine,
		},
	}

	err = mapstructure.Decode(in.KValues, &subsc)
	subsc.Status = subsc.checkSubscription()

	// setup subscription.
	subsc.setup()

	return &subsc, errors.Wrap(err, "create subscription failed")
}

// GetMode returns subscription mode.
func (s *subscription) GetMode() string {
	return s.Mode
}

// InvokeMsg dispose subscription input messages.
func (s *subscription) InvokeMsg() {
	for {
		var msgCtx Message
		if msgCtx = s.mailBox.Get(); nil == msgCtx {
			// detach this entity.
			if atomic.CompareAndSwapInt32(&s.attached, EntityAttached, EntityDetached) {
				log.Infof("detached entity, id: %s.", s.ID)
				break
			}
		}

		// lock messages.
		s.lock.Lock()

		switch msg := msgCtx.(type) {
		case *EntityMessage:
			s.invokeMsg(msg)
		case *TentacleMsg:
			// todo nothing...
		default:
			// invalid msg type.
			log.Errorf("undefine message type, msg: %s", msg)
		}

		s.lock.Unlock()
	}
}

// invokeMsg invoke property messages.
func (s *subscription) invokeMsg(msg *EntityMessage) {
	var err error
	switch s.Mode {
	case SubscriptionModeRealtime:
		err = s.invokeRealtime(msg)
	case SubscriptionModePeriod:
		err = s.invokePeriod(msg)
	case SubscriptionModeChanged:
		err = s.invokeChanged(msg)
	default:
		// invalid subscription mode.
	}

	log.Infof("invoke message: %v,  err: %v", msg, err)
}

// invokeRealtime invoke property where mode is realtime.
func (s *subscription) invokeRealtime(msg *EntityMessage) error {
	// 对于 Realtime 直接转发就OK了.
	bytes, _ := json.Marshal(msg.Values)
	err := s.entityManager.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes)
	return errors.Wrap(err, "invoke realtime message failed")
}

// invokePeriod.
func (s *subscription) invokePeriod(msg *EntityMessage) error {
	// 对于 Period 直接查询快照.
	bytes, _ := json.Marshal(msg.Values)
	err := s.entityManager.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes)
	return errors.Wrap(err, "invoke period message failed")
}

// invokeChanged.
func (s *subscription) invokeChanged(msg *EntityMessage) error {
	// 对于 Changed 直接转发就OK了.
	bytes, _ := json.Marshal(msg.Values)
	err := s.entityManager.daprClient.PublishEvent(context.Background(), s.PubsubName, s.Topic, bytes)
	return errors.Wrap(err, "invoke changed message failed")
}

// checkSubscription returns subscription status.
func (s *subscription) checkSubscription() string {
	if s.Mode == SubscriptionModeUndefine || s.Source == "" ||
		s.Target == "" || s.Filter == "" || s.Topic == "" || s.PubsubName == "" {
		return EntityStatusInactive
	}

	return EntityStatusActive
}

func (s *subscription) setup() error {
	m := mapper.NewMapper(s.ID+"#"+"subscription", s.Filter)

	s.mappers[m.ID()] = m

	// generate indexTentacles again.
	for _, mp := range s.mappers {
		for _, tentacle := range mp.Tentacles() {
			s.indexTentacles[tentacle.TargetID()] =
				append(s.indexTentacles[tentacle.TargetID()], tentacle)
		}
	}

	delete(s.indexTentacles, m.ID())

	log.Info("setup subscription", s.indexTentacles)

	// generate tentacles again.
	s.generateTentacles()

	sourceEntities := []string{}
	for _, expr := range m.SourceEntities() {
		sourceEntities = append(sourceEntities,
			s.entityManager.EscapedEntities(expr)...)
	}

	for _, entityID := range sourceEntities {
		tentacle := mapper.MergeTentacles(s.indexTentacles[entityID]...)

		if nil != tentacle {
			// send tentacle msg.
			s.entityManager.SendMsg(EntityContext{
				Headers: Header{
					EntityCtxHeaderSourceID: s.ID,
					EntityCtxHeaderTargetID: entityID,
				},
				Message: &TentacleMsg{
					TargetID: s.ID,
					Operator: TentacleOperatorAppend,
					Items:    tentacle.Copy().Items(),
				},
			})
		}
	}
	return nil
}

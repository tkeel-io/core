package entities

import (
	"context"
	"sync/atomic"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	// subscription mode enum.
	SubscriptionModeUndefine = "undefine"
	SubscriptionModeRealtime = "realtime"
	SubscriptionModePeriod   = "period"
	SubscriptionModeChanged  = "changed"

	// subscription required fileds.
	SubscriptionFieldMode   = "mode"
	SubScriptionFieldSource = "source"
	SubscriptionFieldTarget = "target"
	SubscriptionFieldFilter = "filter"
)

// SubscriptionBase subscription basic information.
type SubscriptionBase struct {
	Source string `json:"source" mapstructure:"source"`
	Filter string `json:"filter" mapstructure:"filter"`
	Target string `json:"target" mapstructure:"target"`
	Mode   string `json:"mode" mapstructure:"mode"`
}

// subscription subscription actor based entity.
type subscription struct {
	*entity
	SubscriptionBase `mapstructure:",squash"`

	pubsubName string
	topicName  string
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
			// do nothing...
		default:
			// invalid msg type.
			log.Errorf("undefine message type, msg: %s", msg)
		}

		s.lock.Unlock()
	}
}

// invokeMsg invoke property messages.
func (s *subscription) invokeMsg(msg *EntityMessage) {
	switch s.Mode {
	case SubscriptionModeRealtime:
		s.invokeRealtime(msg)
	case SubscriptionModePeriod:
		s.invokePeriod(msg)
	case SubscriptionModeChanged:
		s.invokeChanged(msg)
	default:
		// invalid subscription mode.
	}
}

// invokeRealtime invoke property where mode is realtime.
func (s *subscription) invokeRealtime(msg *EntityMessage) {
	// 对于 Realtime 直接转发就OK了.
	s.entityManager.daprClient.PublishEvent(context.Background(), s.pubsubName, s.topicName, nil)
}

// invokePeriod.
func (s *subscription) invokePeriod(msg *EntityMessage) {
	// 对于 Period 直接查询快照.
	s.entityManager.daprClient.PublishEvent(context.Background(), s.pubsubName, s.topicName, nil)
}

// invokeChanged.
func (s *subscription) invokeChanged(msg *EntityMessage) {
	// 对于 Changed 直接转发就OK了.
	s.entityManager.daprClient.PublishEvent(context.Background(), s.pubsubName, s.topicName, nil)
}

// checkSubscription returns subscription status.
func (s *subscription) checkSubscription() string {
	if s.Mode == SubscriptionModeUndefine || s.Source == "" || s.Target == "" || s.Filter == "" {
		return EntityStatusInactive
	}
	return EntityStatusActive
}

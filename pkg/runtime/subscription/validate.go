package subscription

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/tdtl"
)

// Base subscription basic information.
type Base struct {
	Mode       string `json:"mode" mapstructure:"mode"`
	Filter     string `json:"filter" mapstructure:"filter"`
	Topic      string `json:"topic" mapstructure:"topic"`
	PubsubName string `json:"pubsub_name" mapstructure:"pubsub_name"`
}

// checkSubscription returns subscription status.
func (s *Base) Validate() error {
	// check filter.
	if s.Filter == "" {
		return errors.Wrap(ErrSubscriptionInvalid, "required field filter")
	} else if s.PubsubName == "" {
		return errors.Wrap(ErrSubscriptionInvalid, "required field pubsub_name")
	} else if s.Topic == "" {
		return errors.Wrap(ErrSubscriptionInvalid, "required field topic")
	}

	// check mode.
	switch s.Mode {
	case SubscriptionModeRealtime:
	case SubscriptionModePeriod:
	case SubscriptionModeChanged:
	default:
		return errors.Wrap(ErrSubscriptionInvalid, "subscription mode invalid")
	}

	return nil
}

func decode2Subscription(kvalues map[string]tdtl.Node, subsc *Base) {
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

package runtime

import (
	"context"

	daprSDK "github.com/dapr/go-sdk/client"
	v1 "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/util/dapr"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type SubscriptionMode string

func (sm SubscriptionMode) S() string {
	return string(sm)
}

const (
	SModePeriod    SubscriptionMode = "PERIOD"
	SModeRealtime  SubscriptionMode = "REALTIME"
	SModeOnChanged SubscriptionMode = "ONCHANGED"
)

// 为了订阅实体实现的外部订阅.
func (r *Runtime) handleSubscribe(ctx context.Context, feed *Feed) *Feed {
	log.L().Debug("handle subscribe", zfield.Eid(feed.EntityID), zfield.Event(feed.Event))
	ev, _ := feed.Event.(v1.PatchEvent)

	var err error
	subID := ev.Entity()
	entityID := ev.Attr(v1.MetaSender)
	state, err := r.LoadEntity(subID)
	if nil != err {
		log.L().Error("load entity", zap.Error(err), zfield.Eid(subID))
		feed.Err = err
		return feed
	}

	switch state.Type() {
	case dao.EntityTypeSubscription:
		mode := state.GetProp("mode").String()
		topic := state.GetProp("topic").String()
		pubsubName := state.GetProp("pubsub_name").String()
		log.L().Debug("publish subscription message", zfield.ID(subID), zfield.Event(ev),
			zfield.Eid(entityID), zfield.Topic(topic), zfield.Pubsub(pubsubName), zfield.Mode(mode))

		changes := feed.Patches
		if len(changes) == 0 {
			log.Warn("publish empty message", zfield.ID(subID), zfield.Event(ev),
				zfield.Eid(entityID), zfield.Topic(topic), zfield.Pubsub(pubsubName), zfield.Mode(mode))
			return feed
		}

		var payload []byte
		if payload, err = makePayload(ev, changes); nil != err {
			log.Error("publish message, make payload", zfield.ID(subID), zfield.Event(ev),
				zfield.Eid(entityID), zfield.Topic(topic), zfield.Pubsub(pubsubName), zfield.Mode(mode))
			return feed
		}

		switch mode {
		case SModeRealtime.S():
			ctOpts := daprSDK.PublishEventWithContentType("application/json")
			err = dapr.Get().Select().PublishEvent(ctx, pubsubName, topic, payload, ctOpts)
			if nil != err {
				log.Error("publish message via dapr", zfield.ID(subID), zfield.Event(ev),
					zfield.Eid(entityID), zfield.Topic(topic), zfield.Pubsub(pubsubName), zfield.Mode(mode))
				return feed
			}
		case SModeOnChanged.S():
		//	dapr.Get().Select().PublishEvent(ctx, pubsubName, topic, nil)
		case SModePeriod.S():
		default:
		}
	default:
		return feed
	}

	return feed
}

func makePayload(ev v1.PatchEvent, changes []Patch) ([]byte, error) {
	return []byte(`{}`), nil
}

package runtime

import (
	"context"
	daprSDK "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/util/dapr"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
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

func (r *Runtime) handleSubscribe(ctx context.Context, feed *Feed) *Feed {
	log.L().Debug("handle external subscribe", logf.Eid(feed.EntityID), logf.Event(feed.Event))

	entityID := feed.EntityID
	state := makeSubData(feed)
	if subs, ok := r.entitySubscriptions[entityID]; ok {
		for _, sub := range subs {
			ctOpts := daprSDK.PublishEventWithContentType("application/json")
			err := dapr.Get().Select().PublishEvent(ctx, sub.PubsubName, sub.Topic, state, ctOpts)
			if nil != err {
				log.L().Error("publish message via dapr", logf.ID(sub.ID),
					logf.Eid(entityID), logf.Topic(sub.Topic), logf.Pubsub(sub.PubsubName), logf.Mode(sub.Mode))
				return feed
			}
		}
	}
	return feed
}

func makeSubData(feed *Feed) []byte {
	ret := tdtl.New(`{}`)
	cc := tdtl.New(feed.State)
	for _, change := range feed.Changes {
		path := change.Path
		ret.Set(path, cc.Get(path))
	}
	return ret.Raw()
}

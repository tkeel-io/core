package runtime

import (
	"context"
	"strings"

	daprSDK "github.com/dapr/go-sdk/client"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/metrics"
	"github.com/tkeel-io/core/pkg/repository"
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
	if subs, ok := r.entitySubscriptions[entityID]; ok {
		for _, sub := range subs {
			state := makeSubData(feed, sub)
			if state == nil {
				continue
			}
			metrics.CollectorMsgCount.WithLabelValues(sub.Owner, metrics.MsgTypeSubscribe).Inc()
			log.L().Debug("handle external subs", logf.Eid(feed.EntityID), logf.Event(feed.Event), logf.Any("sub", sub.Filter))
			ctOpts := daprSDK.PublishEventWithContentType("application/json")
			err := dapr.Get().Select().PublishEvent(ctx, sub.PubsubName, sub.Topic, state, ctOpts)
			if nil != err {
				log.L().Error("publish message via dapr", logf.ID(sub.ID),
					logf.Eid(entityID), logf.Topic(sub.Topic), logf.Pubsub(sub.PubsubName), logf.Mode(sub.Mode))
				return feed
			}
		}
	} else {
		log.L().Info("handle external subscribe nil", logf.Eid(feed.EntityID))
	}
	return feed
}

func pathMatch(paths []string, pathCheck string) bool {
	log.L().Info("pathMatch", logf.Any("paths", paths), logf.String("pathCheck", pathCheck))
	for _, path := range paths {
		path = strings.TrimSuffix(path, "*")
		if strings.HasPrefix(pathCheck, path) {
			return true
		}
		continue
	}
	return false
}

func makeSubData(feed *Feed, sub *repository.Subscription) []byte {
	ret := tdtl.New(`{}`)
	cc := tdtl.New(feed.State)
	writeFlag := false
	for _, change := range feed.Changes {
		path := change.Path
		if pathMatch(sub.SourceEntityPaths, path) {
			ret.Set(path, cc.Get(path))
			writeFlag = true
		}
	}
	if !writeFlag {
		return nil
	}

	ret.Set("id", tdtl.NewString(feed.EntityID))
	ret.Set("subscribe_id", tdtl.NewString(sub.ID))
	ret.Set("owner", tdtl.NewString(sub.Owner))

	return ret.Raw()
}

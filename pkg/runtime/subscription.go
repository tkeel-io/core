package runtime

import (
	"context"

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
	ev, _ := feed.Event.(v1.PatchEvent)
	entityID := ev.Entity()

	state, err := r.LoadEntity(entityID)
	if nil != err {
		log.L().Error("load entity", zap.Error(err), zfield.Eid(entityID))
		feed.Err = err
		return feed
	}

	switch state.Type() {
	case dao.EntityTypeSubscription:
		mode := state.GetProp("mode").String()
		topic := state.GetProp("topic").String()
		pubsubName := state.GetProp("pubsub_name").String()
		switch mode {
		case SModeRealtime.S():
			dapr.Get().Select().PublishEvent(ctx, pubsubName, topic, []byte(`{}`))
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

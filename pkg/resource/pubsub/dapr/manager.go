package dapr

import (
	"context"

	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/kit/log"
)

// 场景：
// 由于订阅需要通过启动边车时配置 component.Pubsub 和 component.Subscription 对象， 不能动态配置.
// 我们所有的 pubsub 订阅使用统一的 component.Pubsub 和 component.Subscription.
// 所以我们需要在 边车订阅的基础上将 订阅的消息分离.
// warn: 但是这可能会导致没有被处理的消息丢失.
// warn: consumer 的销毁需要处理.

const (
	// SubscriptionResponseStatusSuccess means message is processed successfully.
	SubscriptionResponseStatusSuccess = "SUCCESS"
	// SubscriptionResponseStatusRetry means message to be retried by Dapr.
	SubscriptionResponseStatusRetry = "RETRY"
	// SubscriptionResponseStatusDrop means warning is logged and message is dropped.
	SubscriptionResponseStatusDrop = "DROP"
)

var (
	clusterConsumer = defaultConsumer
)

func Register(consumer *Consumer) {
	clusterConsumer = consumer
}

func Unregister(consumer *Consumer) {
	clusterConsumer = defaultConsumer
}

func HandleEvent(ctx context.Context, ev v1.Event) (out *v1.TopicEventResponse, err error) {
	// dispatch message.
	err = clusterConsumer.handler(ctx, ev)
	return &v1.TopicEventResponse{Status: SubscriptionResponseStatusSuccess}, errors.Wrap(err, "handle event")
}

type Consumer struct {
	id      string
	handler pubsub.EventHandler
}

var defaultConsumer = &Consumer{id: "defaultConsumer", handler: func(ctx context.Context, e v1.Event) error {
	log.L().Warn("empty cluster consumer", logf.ID(e.ID()), logf.Any("event", e))
	return nil
}}

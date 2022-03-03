package dapr

import (
	"context"
	"fmt"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go"
	pb "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/message"
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
	lock      = sync.RWMutex{}
	consumers map[string][]*Consumer
)

type Consumer struct {
	id      string
	handler pubsub.EventHandler
}

func HandleEvent(ctx context.Context, ev cloudevents.Event) (out *pb.TopicEventResponse, err error) {
	var topic string
	var pubsubName string
	ev.ExtensionAs(message.ExtCloudEventTopic, &topic)
	ev.ExtensionAs(message.ExtCloudEventPubsub, &pubsubName)

	// dispatch message.
	groupName := consumerGroup(pubsubName, topic)
	log.Debug("handle event", zfield.Topic(topic),
		zfield.Header(message.GetAttributes(ev)), zfield.Pubsub(pubsubName))

	lock.RLock()
	for _, consumer := range consumers[groupName] {
		consumer.handler(ctx, ev)
	}
	lock.RUnlock()

	return &pb.TopicEventResponse{Status: SubscriptionResponseStatusSuccess}, nil
}

func consumerGroup(pubsubName, topicName string) string {
	return fmt.Sprintf("%s/%s", pubsubName, topicName)
}

func registerConsumer(pubsubName, topicName string, consumer *Consumer) error {
	lock.Lock()
	defer lock.Unlock()
	group := consumerGroup(pubsubName, topicName)
	consumers[group] = append(consumers[group], consumer)
	return nil
}

func unregisterConsumer(pubsubName, topicName string, consumer *Consumer) error {
	lock.Lock()
	defer lock.Unlock()
	groupName := consumerGroup(pubsubName, topicName)
	consumerGrop := consumers[groupName]
	for index, cs := range consumerGrop {
		if cs.id == consumer.id {
			consumers[groupName] =
				append(consumerGrop[:index], consumerGrop[index+1:]...)
			break
		}
	}
	return nil
}

func init() {
	consumers = make(map[string][]*Consumer)
}

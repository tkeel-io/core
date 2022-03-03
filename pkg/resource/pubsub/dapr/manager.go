package dapr

import (
	"context"
	"fmt"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
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
	consumerManager *ConsumerManager
)

func Get() *ConsumerManager {
	return consumerManager
}

type ConsumerManager struct {
	lock            sync.RWMutex
	clusterConsumer *Consumer
	nodeConsumers   map[string][]*Consumer
}

func (cm *ConsumerManager) Register(consumerType, pubsubName, topicName string, consumer *Consumer) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	switch dao.ConsumerType(consumerType) {
	case dao.ConsumerTypeDispatch:
		cm.clusterConsumer = consumer
	case dao.ConsumerTypeCore:
		group := consumerGroup(pubsubName, topicName)
		cm.nodeConsumers[group] = append(cm.nodeConsumers[group], consumer)
	default:
		return xerrors.ErrInvalidQueueConsumerType
	}

	return nil
}

func (cm *ConsumerManager) Unregister(consumerType, pubsubName, topicName string, consumer *Consumer) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	switch dao.ConsumerType(consumerType) {
	case dao.ConsumerTypeDispatch:
		cm.clusterConsumer = defaultConsumer
	case dao.ConsumerTypeCore:
		groupName := consumerGroup(pubsubName, topicName)
		consumerGrop := cm.nodeConsumers[groupName]
		for index, cs := range consumerGrop {
			if cs.id == consumer.id {
				cm.nodeConsumers[groupName] =
					append(consumerGrop[:index], consumerGrop[index+1:]...)
				break
			}
		}
	default:
		return xerrors.ErrInvalidQueueConsumerType
	}

	return nil
}

func (cm *ConsumerManager) DeliveredEvent(ctx context.Context, ev cloudevents.Event) (out *pb.TopicEventResponse, err error) {
	var topic string
	var pubsubName string
	var consumerType string
	ev.ExtensionAs(message.ExtCloudEventTopic, &topic)
	ev.ExtensionAs(message.ExtCloudEventPubsub, &pubsubName)
	ev.ExtensionAs(message.ExtCloudEventConsumerType, &consumerType)

	// dispatch message.
	elapsedTime := util.NewElapsed()
	log.Debug("handle event", zfield.Topic(topic),
		zfield.Header(message.GetAttributes(ev)), zfield.Pubsub(pubsubName))

	cm.lock.RLock()
	handlers := make([]pubsub.EventHandler, 0)
	switch dao.ConsumerType(consumerType) {
	case dao.ConsumerTypeDispatch:
		handlers = append(handlers, cm.clusterConsumer.handler)
	case dao.ConsumerTypeCore:
		groupName := consumerGroup(pubsubName, topic)
		for _, consumer := range cm.nodeConsumers[groupName] {
			handlers = append(handlers, consumer.handler)
		}
	default:
		log.Error("handle event", zfield.Topic(topic), zap.Error(err),
			zfield.Header(message.GetAttributes(ev)), zfield.Pubsub(pubsubName))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, nil
	}
	cm.lock.RUnlock()

	// dispose event.
	for _, handler := range handlers {
		if err = handler(ctx, ev); nil != err {
			log.Error("handle event", zfield.Topic(topic), zap.Error(err),
				zfield.Header(message.GetAttributes(ev)), zfield.Pubsub(pubsubName))
		}
	}

	log.Debug("handle event completed", zfield.Topic(topic),
		zfield.Elapsedms(time.Duration(elapsedTime.ElapsedMilli())),
		zfield.Header(message.GetAttributes(ev)), zfield.Pubsub(pubsubName))
	return &pb.TopicEventResponse{Status: SubscriptionResponseStatusSuccess}, nil
}

type Consumer struct {
	id      string
	handler pubsub.EventHandler
}

var defaultConsumer = &Consumer{id: "defaultConsumer", handler: func(ctx context.Context, e cloudevents.Event) error {
	log.Warn("empty cluster consumer", zfield.Header(message.GetAttributes(e)), zfield.Event(e))
	return nil
}}

func consumerGroup(pubsubName, topicName string) string {
	return fmt.Sprintf("%s/%s", pubsubName, topicName)
}

func init() {
	consumerManager = &ConsumerManager{
		lock:          sync.RWMutex{},
		nodeConsumers: make(map[string][]*Consumer),
	}
}

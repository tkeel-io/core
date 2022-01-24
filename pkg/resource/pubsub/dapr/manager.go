package dapr

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/message"
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
	lock      = sync.RWMutex{}
	consumers map[string][]*Consumer
)

type Consumer struct {
	id      string
	handler pubsub.MessageHandler
}

func HandleEvent(ctx context.Context, req *pb.TopicEventRequest) (out *pb.TopicEventResponse, err error) {
	var values map[string]interface{}
	var properties map[string]constraint.Node
	switch kv := req.Data.AsInterface().(type) {
	case map[string]interface{}:
		values = kv

	default:
		log.Warn("invalid event", zap.String("id", req.Id), zap.Any("event", req))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, nil
	}

	// parse data.
	switch data := values["data"].(type) {
	case map[string]interface{}:
		if len(data) > 0 {
			properties = make(map[string]constraint.Node)
			for key, val := range data {
				properties[key] = constraint.NewNode(val)
			}
		}
	default:
		log.Warn("invalid event", zap.String("id", req.Id), zap.Any("event", req))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, nil
	}

	msgCtx := message.MessageContext{
		Headers: message.Header{},
		Message: message.PropertyMessage{
			StateID:    interface2string(values["id"]),
			Operator:   constraint.PatchOpReplace.String(),
			Properties: properties,
		},
	}

	msgCtx.Headers.SetReceiver(interface2string(values["id"]))
	msgCtx.Headers.SetOwner(interface2string(values["owner"]))
	msgCtx.Headers.SetOwner(interface2string(values["type"]))
	msgCtx.Headers.SetOwner(interface2string(values["source"]))

	// dispatch message.
	groupName := consumerGroup(req.Pubsubname, req.Topic)

	lock.RLock()
	for _, consumer := range consumers[groupName] {
		consumer.handler(ctx, msgCtx)
	}
	lock.RUnlock()

	return &pb.TopicEventResponse{Status: SubscriptionResponseStatusSuccess}, nil
}

func interface2string(in interface{}) (out string) {
	switch inString := in.(type) {
	case string:
		out = inString
	case constraint.Node:
		out = inString.String()
	default:
		out = ""
	}
	return
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

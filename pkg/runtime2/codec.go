package runtime2

import "github.com/Shopify/sarama"

func deliveredEvent(msg *sarama.ConsumerMessage) *ContainerEvent {
	ev := &ContainerEvent{}
	return ev
}

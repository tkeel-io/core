package runtime3

import "github.com/Shopify/sarama"

func deliveredEvent(msg *sarama.ConsumerMessage) *ContainerEvent {
	ev := &ContainerEvent{}
	return ev
}

type EventType string
type ContainerEventType string

const (
	OpContainer    EventType          = "core.operation.Container"
	OpEntity       EventType          = "core.operation.Entity"
	OpCache        EventType          = "core.operation.Cache"
	OpMapperCreate ContainerEventType = "core.operation.Mapper.Create"
	OpMapperUpdate ContainerEventType = "core.operation.Mapper.Update"
	OpMapperDelete ContainerEventType = "core.operation.Mapper.Delete"
	OpEntityCreate ContainerEventType = "core.operation.Entity.Create"
	OpEntityUpdate ContainerEventType = "core.operation.Entity.Update"
	OpEntityDelete ContainerEventType = "core.operation.Entity.Delete"
)

//处理消息
type ContainerEvent struct {
	ID       string
	Type     EventType
	Callback string
	Value    interface{}
}

package message

type Message interface {
	Type() MessageType
}

type MessageType string //nolint

func (mt MessageType) String() string {
	return string(mt)
}

const (
	MessageTypeState        MessageType = "state"
	MessageTypeAPIRequest   MessageType = "apirequest"
	MessageTypeAPIRespond   MessageType = "apirespond"
	MessageTypeAPIRepublish MessageType = "republish"
)

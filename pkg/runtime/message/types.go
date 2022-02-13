package message

type Message interface {
	Type() MessageType
}

type MessageType string //nolint

func (mt MessageType) String() string {
	return string(mt)
}

const (
	MessageTypeSync    MessageType = "sync"
	MessageTypeState   MessageType = "state"
	MessageTypeRespond MessageType = "respond"
)

package message

type Message interface {
	Type() MessageType
}

type MessageType string //nolint

func (mt MessageType) String() string {
	return string(mt)
}

const (
	MessageTypeRaw          MessageType = "raw"
	MessageTypeState        MessageType = "state"
	MessageTypeMapperInit   MessageType = "mapper"
	MessageTypeAPIRequest   MessageType = "apirequest"
	MessageTypeAPIRespond   MessageType = "apirespond"
	MessageTypeAPIRepublish MessageType = "republish"
)

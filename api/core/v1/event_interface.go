package v1

type Attribution interface {
	Attr(key string) string
	SetAttr(key string, value string) Message
	ForeachAttr(handler func(key, val string) error) error
}

type Message interface {
	Attribution
	Copy() Message
	Type() string
	Version() string
	Validate() error
	Entity() string
	SetEntity(entityId string) Message
	SetTTL(td int) Message

	Data() []byte
	SetData(data []byte) Message
}

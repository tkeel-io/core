package message

const (
	MsgCtxHeaderType      = "x-type"
	MsgCtxHeaderOwner     = "x-owner"
	MsgCtxHeaderSource    = "x-source"
	MsgCtxHeaderSender    = "x-sender"
	MsgCtxHeaderReceiver  = "x-receiver"
	MsgCtxHeaderRequestID = "x-reqsuest-id"
	MsgCtxHeaderMessageID = "x-message-id"
	MsgCtxHeaderChannelID = "x-channel-id"
	MsgCtxHeaderTemplate  = "x-template-id"
)

type Header map[string]string

type MessageContext struct { //nolint
	Headers Header
	Message Message
}

func (h Header) Get(key string) string       { return h[key] }
func (h Header) Set(key, value string)       { h[key] = value }
func (h Header) GetType() string             { return h[MsgCtxHeaderType] }
func (h Header) SetType(typ string)          { h[MsgCtxHeaderType] = typ }
func (h Header) GetOwner() string            { return h[MsgCtxHeaderOwner] }
func (h Header) SetOwner(owner string)       { h[MsgCtxHeaderOwner] = owner }
func (h Header) GetSource() string           { return h[MsgCtxHeaderSource] }
func (h Header) SetSource(source string)     { h[MsgCtxHeaderSource] = source }
func (h Header) GetSender() string           { return h[MsgCtxHeaderSender] }
func (h Header) SetSender(sender string)     { h[MsgCtxHeaderSender] = sender }
func (h Header) GetReceiver() string         { return h[MsgCtxHeaderReceiver] }
func (h Header) SetReceiver(receiver string) { h[MsgCtxHeaderReceiver] = receiver }
func (h Header) GetTemplate() string         { return h[MsgCtxHeaderTemplate] }
func (h Header) SetTemplate(template string) { h[MsgCtxHeaderTemplate] = template }
func (h Header) GetRequestID() string        { return h[MsgCtxHeaderRequestID] }
func (h Header) SetRequestID(reqID string)   { h[MsgCtxHeaderRequestID] = reqID }
func (h Header) GetMessageID() string        { return h[MsgCtxHeaderMessageID] }
func (h Header) SetMessageID(msgID string)   { h[MsgCtxHeaderMessageID] = msgID }

func (h Header) GetDefault(key, defaultValue string) string {
	if _, has := h[key]; !has {
		return defaultValue
	}
	return h[key]
}

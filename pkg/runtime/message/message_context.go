package message

const (
	MsgCtxHeaderType         = "x-type"
	MsgCtxHeaderOwner        = "x-owner"
	MsgCtxHeaderSource       = "x-source"
	MsgCtxHeaderSender       = "x-sender"
	MsgCtxHeaderReceiver     = "x-receiver"
	MsgCtxHeaderRequestID    = "x-reqsuest-id"
	MsgCtxHeaderMessageID    = "x-message-id"
	MsgCtxHeaderChannelID    = "x-channel-id"
	MsgCtxHeaderTemplate     = "x-template-id"
	MsgCtxHeaderMsgFlush     = "x-flush-flag"
	MsgCtxHeaderMMessageType = "x-message-type"
)

type Header map[string]string

type MessageContext struct { // nolint
	Headers Header
	Message Message
	RawData string
}

func NewMessageContext(msg Message) MessageContext {
	return MessageContext{
		Headers: make(Header),
		Message: msg,
	}
}

func (msgCtx MessageContext) Get(key string) string   { return msgCtx.Headers[key] }
func (msgCtx MessageContext) Set(key, value string)   { msgCtx.Headers[key] = value }
func (msgCtx MessageContext) GetRaw() []byte          { return []byte(msgCtx.RawData) }
func (msgCtx MessageContext) GetType() string         { return msgCtx.Headers[MsgCtxHeaderType] }
func (msgCtx MessageContext) SetType(typ string)      { msgCtx.Headers[MsgCtxHeaderType] = typ }
func (msgCtx MessageContext) GetOwner() string        { return msgCtx.Headers[MsgCtxHeaderOwner] }
func (msgCtx MessageContext) SetOwner(owner string)   { msgCtx.Headers[MsgCtxHeaderOwner] = owner }
func (msgCtx MessageContext) GetSource() string       { return msgCtx.Headers[MsgCtxHeaderSource] }
func (msgCtx MessageContext) SetSource(source string) { msgCtx.Headers[MsgCtxHeaderSource] = source }
func (msgCtx MessageContext) GetSender() string       { return msgCtx.Headers[MsgCtxHeaderSender] }
func (msgCtx MessageContext) SetSender(sender string) { msgCtx.Headers[MsgCtxHeaderSender] = sender }
func (msgCtx MessageContext) GetReceiver() string     { return msgCtx.Headers[MsgCtxHeaderReceiver] }
func (msgCtx MessageContext) SetMsgType(msgType string) {
	msgCtx.Headers[MsgCtxHeaderMMessageType] = msgType
}
func (msgCtx MessageContext) GetMsgType() string { return msgCtx.Headers[MsgCtxHeaderMMessageType] }
func (msgCtx MessageContext) GetHeader() Header  { return msgCtx.Headers }

func (msgCtx MessageContext) SetReceiver(receiver string) {
	msgCtx.Headers[MsgCtxHeaderReceiver] = receiver
}

func (msgCtx MessageContext) GetTemplate() string { return msgCtx.Headers[MsgCtxHeaderTemplate] }

func (msgCtx MessageContext) SetTemplate(template string) {
	msgCtx.Headers[MsgCtxHeaderTemplate] = template
}

func (msgCtx MessageContext) GetRequestID() string { return msgCtx.Headers[MsgCtxHeaderRequestID] }
func (msgCtx MessageContext) SetRequestID(reqID string) {
	msgCtx.Headers[MsgCtxHeaderRequestID] = reqID
}
func (msgCtx MessageContext) GetMessageID() string { return msgCtx.Headers[MsgCtxHeaderMessageID] }
func (msgCtx MessageContext) SetMessageID(msgID string) {
	msgCtx.Headers[MsgCtxHeaderMessageID] = msgID
}

func (msgCtx MessageContext) GetDefault(key, defaultValue string) string {
	if _, has := msgCtx.Headers[key]; !has {
		return defaultValue
	}
	return msgCtx.Headers[key]
}

package message

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

const (
	// event extension fields definitions.
	ExtEntityID        = "extenid"
	ExtEntityType      = "extentype"
	ExtEntityOwner     = "extowner"
	ExtEntitySource    = "extsource"
	ExtTemplateID      = "exttemplate"
	ExtMessageID       = "extmsgid"
	ExtMessageType     = "extmsgtype"
	ExtMessageSender   = "extsender"
	ExtMessageReceiver = "extreceiver"
	ExtChannelID       = "extchid"
	ExtCallback        = "extcallback"
	ExtAPIIdentify     = "extapiid"

	ExtCloudEventID          = "exteventid"
	ExtCloudEventSpec        = "exteventspec"
	ExtCloudEventType        = "exteventtype"
	ExtCloudEventSource      = "exteventsource"
	ExtCloudEventSubject     = "exteventsubject"
	ExtCloudEventDataSchema  = "exteventschema"
	ExtCloudEventContentType = "exteventcontenttype"
)

/*

定义消息：
	1. 来自 api 的消息 event.Data 中包含 api 调用的参数.
	2. 状态的消息，定义为 满足 json.Json.
*/

func GetAttributes(event cloudevents.Event) map[string]string {
	var attributes = make(map[string]string)
	// construct attributes from CloudEvent.
	attributes[ExtCloudEventID] = event.ID()
	attributes[ExtCloudEventSpec] = event.SpecVersion()
	attributes[ExtCloudEventType] = event.Type()
	attributes[ExtCloudEventSource] = event.Source()
	attributes[ExtCloudEventSubject] = event.Subject()
	attributes[ExtCloudEventDataSchema] = event.DataSchema()
	attributes[ExtCloudEventContentType] = event.DataContentType()
	for key, val := range event.Extensions() {
		if value, ok := val.(string); ok {
			attributes[key] = value
			continue
		}
		log.Warn("missing attributes field", zfield.Key(key), zfield.Value(val))
	}
	return attributes
}

type Context struct {
	attributes map[string]string
	data       []byte
	ctx        context.Context
}

func New(ctx context.Context) Context {
	return Context{
		ctx:        ctx,
		attributes: make(map[string]string),
	}
}

func From(ctx context.Context, ev cloudevents.Event) (Context, error) {
	bytes, err := ev.DataBytes()
	if nil != err {
		log.Error("parse event", zfield.ID(ev.ID()), zfield.Header(GetAttributes(ev)))
	}

	msgCtx := Context{
		ctx:        ctx,
		data:       bytes,
		attributes: GetAttributes(ev),
	}

	return msgCtx, errors.Wrap(err, "parse event")
}

func (ctx *Context) value(key string) interface{} {
	if val, ok := ctx.attributes[key]; ok {
		return val
	}

	// check context.
	return ctx.ctx.Value(key)
}

func (ctx *Context) Get(key string) string {
	val := ctx.value(key)
	valStr, _ := val.(string)
	return valStr
}

func (ctx *Context) Set(key string, val string) {
	ctx.attributes[key] = val
}

func (ctx *Context) Message() []byte {
	return ctx.data
}

func (ctx *Context) Context() context.Context {
	return ctx.ctx
}

func (ctx *Context) Attributes() map[string]string {
	return ctx.attributes
}

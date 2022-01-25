package message

import (
	"context"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
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
	ExtRequestID       = "extreqid"
	ExtChannelID       = "extchid"
	ExtPromise         = "extpromise"
	ExtSyncFlag        = "extsync"

	ExtCloudEventID          = "exteventid"
	ExtCloudEventSpec        = "exteventspec"
	ExtCloudEventType        = "exteventtype"
	ExtCloudEventSource      = "exteventsource"
	ExtCloudEventSubject     = "exteventsubject"
	ExtCloudEventDataSchema  = "exteventschema"
	ExtCloudEventContentType = "exteventcontenttype"

	ExtValueSync = "sync"
)

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
		}
		log.Warn("missing attributes field", zfield.Key(key), zfield.Value(val))
	}
	return attributes
}

type Context struct {
	waiter     util.Waiter
	attributes map[string]string
	message    Message
	ctx        context.Context
}

func New(ctx context.Context) Context {
	return Context{
		ctx:        ctx,
		waiter:     util.NewWaiter(),
		attributes: make(map[string]string),
	}
}

func From(ctx context.Context, ev cloudevents.Event) (Context, error) {
	msgCtx := Context{ctx: ctx}
	waiter := util.NewWaiter()
	attributes := GetAttributes(ev)
	if ExtValueSync == attributes[ExtSyncFlag] {
		waiter = &sync.WaitGroup{}
		waiter.Add(1)
	}

	var err error
	var msgType MessageType
	ev.ExtensionAs(ExtMessageType, &msgType)
	switch msgType {
	case MessageTypeState:
		var msg StateMessage
		if err = ev.DataAs(&msg); nil != err {
			log.Error("parse state message", zap.Error(err), zfield.Event(ev))
			return msgCtx, errors.Wrap(err, "parse state message")
		}
		// set promise handler.
		msg.MessageBase = NewBase(func(v interface{}) {
			log.Debug("process message successed")
			waiter.Done()
		})

		msgCtx.message = msg
	case MessageTypeProps:
		var rawData []byte
		if rawData, err = ev.DataBytes(); nil != err {
			log.Error("parse props message", zap.Error(err), zfield.Event(ev))
			return msgCtx, errors.Wrap(err, "parse props message")
		}

		// decode property message.
		msg, err := defaultPropsCodec.Decode(rawData)
		if nil != err {
			log.Error("decode props message", zap.Error(err), zfield.Event(ev))
			return msgCtx, errors.Wrap(err, "decode props message")
		}

		// set promise handler.
		msg.MessageBase = NewBase(func(v interface{}) {
			log.Debug("process message successed")
			waiter.Done()
		})

		msgCtx.message = msg
	default:
		waiter.Done()
		log.Error("invalid message type", zfield.Event(ev))
		return Context{}, xerrors.ErrInvalidMessageType
	}

	return Context{
		ctx:        ctx,
		waiter:     waiter,
		message:    msgCtx.message,
		attributes: attributes,
	}, nil
}

func (ctx *Context) Value(key string) interface{} {
	if val, ok := ctx.attributes[key]; ok {
		return val
	}

	// check context.
	return ctx.ctx.Value(key)
}

func (ctx *Context) Get(key string) string {
	val := ctx.Value(key)
	valStr, _ := val.(string)
	return valStr
}

func (ctx *Context) Set(key string, val string) {
	ctx.attributes[key] = val
}

func (ctx *Context) With(msg Message) {
	ctx.message = msg
}

func (ctx *Context) SetWaiter(waiter util.Waiter) {
	ctx.waiter = waiter
}

func (ctx *Context) Message() Message {
	return ctx.message
}

func (ctx *Context) Wait() {
	ctx.waiter.Wait()
}

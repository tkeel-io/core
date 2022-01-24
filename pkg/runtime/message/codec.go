package message

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
	msgpack "github.com/shamaton/msgpack/v2"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func Encode(msgCtx *MessageContext) error {
	switch msg := msgCtx.Message.(type) {
	case StateMessage:
		bytes, err := json.Marshal(msg)
		if nil != err {
			log.Error("json marshal StateMessage", zap.Error(err), zfield.Message(msgCtx))
			return errors.Wrap(err, "json marshal StateMessage")
		}
		msgCtx.SetMsgType(MessageTypeState)
		msgCtx.RawData = base64.StdEncoding.EncodeToString(bytes)
	case FlushPropertyMessage:
		bytes, err := msgpack.Marshal(msg)
		if nil != err {
			log.Error("json marshal StateMessage", zap.Error(err), zfield.Message(msgCtx))
			return errors.Wrap(err, "json marshal StateMessage")
		}
		msgCtx.SetMsgType(MessageTypeFlushProperty)
		msgCtx.RawData = base64.StdEncoding.EncodeToString(bytes)
	case PropertyMessage:
		bytes, err := msgpack.Marshal(msg)
		if nil != err {
			log.Error("json marshal StateMessage", zap.Error(err), zfield.Message(msgCtx))
			return errors.Wrap(err, "json marshal StateMessage")
		}
		msgCtx.SetMsgType(MessageTypeProperty)
		msgCtx.RawData = base64.StdEncoding.EncodeToString(bytes)
	default:
		return xerrors.ErrMessageTypeInvalid
	}
	return nil
}

func Decode(msgCtx MessageContext) (MessageContext, error) {
	bytes, err := base64.StdEncoding.DecodeString(msgCtx.RawData)
	if nil != err {
		log.Error("base64 unmarshal RawData", zap.Error(err), zfield.Message(msgCtx))
		return msgCtx, errors.Wrap(err, "base64.decode")
	}
	switch msgCtx.GetMsgType() {
	case MessageTypeState:
		var msg StateMessage
		if err = json.Unmarshal(bytes, &msg); nil != err {
			log.Error("json unmarshal StateManager", zap.Error(err), zfield.Message(msgCtx))
		}
		msgCtx.Message = msg
	case MessageTypeProperty:
		var msg PropertyMessage
		msgpack.Unmarshal(bytes, &msg)
		msgCtx.Message = msg
	case MessageTypeFlushProperty:
		var msg PropertyMessage
		msgpack.Unmarshal(bytes, &msg)
		msgCtx.Message = msg
	default:
		return msgCtx, xerrors.ErrMessageTypeInvalid
	}
	return msgCtx, errors.Wrap(err, "decode MessageContext")
}

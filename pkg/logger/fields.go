package logger

import (
	"time"

	"go.uber.org/zap"
)

// Eid returns enitty id field.
func Eid(id string) zap.Field {
	return zap.String("entity_id", id)
}

// Message returns message field.
func Message(msg interface{}) zap.Field {
	return zap.Any("message", msg)
}

// TQL returns TQL field.
func TQL(tql string) zap.Field {
	return zap.String("TQL", tql)
}

// ReqID returns request id field.
func ReqID(reqID string) zap.Field {
	return zap.String("request_id", reqID)
}

// MsgID returns message id field.
func MsgID(msgID string) zap.Field {
	return zap.String("message_id", msgID)
}

// Mid returns mapper id field.
func Mid(id string) zap.Field {
	return zap.String("mapper_id", id)
}

// PK returns property key field.
func PK(key string) zap.Field {
	return zap.String("property_key", key)
}

// Target return target key field.
func Target(target string) zap.Field {
	return zap.String("target", target)
}

// Op returns operator field.
func Op(op string) zap.Field {
	return zap.String("op", op)
}

// Type return type key field.
func Type(t string) zap.Field {
	return zap.String("type", t)
}

// Status returns status key field.
func Status(status string) zap.Field {
	return zap.String("status", status)
}

func Base(base map[string]interface{}) zap.Field {
	return zap.Any("base", base)
}

func ID(id string) zap.Field {
	return zap.String("id", id)
}

func Path(path string) zap.Field {
	return zap.String("path", path)
}

func Elapsed(duration time.Duration) zap.Field {
	return zap.Duration("elapsed", duration)
}

func Elapsedms(duration time.Duration) zap.Field {
	return zap.Duration("elapsedms", time.Duration(duration.Milliseconds()))
}

func Reason(reason string) zap.Field {
	return zap.String("reason", reason)
}

func Owner(owner string) zap.Field {
	return zap.String("owner", owner)
}

func Source(source string) zap.Field {
	return zap.String("source", source)
}

func Template(tid string) zap.Field {
	return zap.String("template_id", tid)
}

func Key(key string) zap.Field {
	return zap.String("key", key)
}

func Value(val interface{}) zap.Field {
	return zap.Any("value", val)
}

func Desc(description string) zap.Field {
	return zap.String("description", description)
}

func Name(name string) zap.Field {
	return zap.String("name", name)
}

func Sender(sender string) zap.Field {
	return zap.String("sender", sender)
}

func Receiver(receiver string) zap.Field {
	return zap.String("receiver", receiver)
}

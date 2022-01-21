package logger

import (
	"time"

	"go.uber.org/zap"
)

func Eid(id string) zap.Field {
	return zap.String("entity_id", id)
}

func Message(msg interface{}) zap.Field {
	return zap.Any("message", msg)
}

func TQL(tql string) zap.Field {
	return zap.String("TQL", tql)
}

func ReqID(reqID string) zap.Field {
	return zap.String("request_id", reqID)
}

func MsgID(msgID string) zap.Field {
	return zap.String("message_id", msgID)
}

func Mid(id string) zap.Field {
	return zap.String("mapper_id", id)
}

func PK(key string) zap.Field {
	return zap.String("property_key", key)
}

func Target(target string) zap.Field {
	return zap.String("target", target)
}

func Op(op string) zap.Field {
	return zap.String("op", op)
}

func Type(t string) zap.Field {
	return zap.String("type", t)
}

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

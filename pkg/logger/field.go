package logger

import "go.uber.org/zap"

func EntityID(id string) zap.Field {
	return zap.String("entity_id", id)
}

func MessageInst(msg interface{}) zap.Field {
	return zap.Any("message", msg)
}

func TQLString(tql string) zap.Field {
	return zap.String("TQL", tql)
}

func RequestID(reqID string) zap.Field {
	return zap.String("request_id", reqID)
}

func MapperID(id string) zap.Field {
	return zap.String("mapper_id", id)
}

func PropertyKey(key string) zap.Field {
	return zap.String("property_key", key)
}

func Target(target string) zap.Field {
	return zap.String("target", target)
}

func Operator(op string) zap.Field {
	return zap.String("op", op)
}

func Type(t string) zap.Field {
	return zap.String("type", t)
}

func Status(status string) zap.Field {
	return zap.String("status", status)
}

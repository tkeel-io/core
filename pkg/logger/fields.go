package logger

import (
	"encoding/json"
	"time"

	v1 "github.com/tkeel-io/core/api/core/v1"
	"go.uber.org/zap"
)

// Eid returns enitty id field.
func Eid(id string) zap.Field {
	return zap.String("entity_id", id)
}

func EvID(id string) zap.Field {
	return zap.String("event_id", id)
}

func RID(id string) zap.Field {
	return zap.String("runtime_id", id)
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

func URL(urlText string) zap.Field {
	return zap.String("url", urlText)
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

func Entity(entityJSON string) zap.Field {
	return zap.Any("entity", entityJSON)
}

func ID(id string) zap.Field {
	return zap.String("id", id)
}

func Channel(channel string) zap.Field {
	return zap.String("channel", channel)
}

func Path(path string) zap.Field {
	return zap.String("path", path)
}

func Elapsed(duration time.Duration) zap.Field {
	return zap.Duration("elapsed", duration)
}

func Elapsedms(duration int64) zap.Field {
	return zap.Int64("elapsedms", duration)
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

func Expr(expr string) zap.Field {
	return zap.String("expression", expr)
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

func Lease(lease int64) zap.Field {
	return zap.Int64("lease", lease)
}

func Cluster(cid uint64) zap.Field {
	return zap.Uint64("cluster", cid)
}

func Member(mid uint64) zap.Field {
	return zap.Uint64("member", mid)
}

func Revision(revision uint64) zap.Field {
	return zap.Uint64("revision", revision)
}

func Term(term int64) zap.Field {
	return zap.Int64("term", term)
}

func Prefix(prefix string) zap.Field {
	return zap.String("prefix", prefix)
}

func Count(n int64) zap.Field {
	return zap.Int64("count", n)
}

func Endpoints(endpoints []string) zap.Field {
	return zap.Strings("endpoints", endpoints)
}

func App(appID string) zap.Field {
	return zap.String("app_id", appID)
}

func Host(host string) zap.Field {
	return zap.String("host", host)
}

func Port(port int) zap.Field {
	return zap.Int("port", port)
}

func Version(version int64) zap.Field {
	return zap.Int64("version", version)
}

func DispatcherID(id string) zap.Field {
	return zap.String("dispatcher_id", id)
}

func DispatcherName(name string) zap.Field {
	return zap.String("dispatcher_name", name)
}

func Mode(mode string) zap.Field {
	return zap.String("mode", mode)
}

func Topic(topic string) zap.Field {
	return zap.String("topic", topic)
}

func Pubsub(pubsub string) zap.Field {
	return zap.String("pubsub", pubsub)
}

func Event(ev v1.Event) zap.Field {
	bytes, _ := json.Marshal(ev)
	return zap.Any("event", string(bytes))
}

func Spec(spec string) zap.Field {
	return zap.String("spec", spec)
}

func Method(method string) zap.Field {
	return zap.String("method", method)
}

func Header(header map[string]string) zap.Field {
	return zap.Any("header", header)
}

func Addr(addr string) zap.Field {
	return zap.String("address", addr)
}

func Payload(payload []byte) zap.Field {
	return zap.String("payload", string(payload))
}

func Partition(partition int32) zap.Field {
	return zap.Int32("partition", partition)
}

func Offset(offset int64) zap.Field {
	return zap.Int64("offset", offset)
}

func Group(group string) zap.Field {
	return zap.String("offset", group)
}

func Queue(v interface{}) zap.Field {
	return zap.Any("queue", v)
}

func Input(v interface{}) zap.Field {
	return zap.Any("input", v)
}

func Output(v interface{}) zap.Field {
	return zap.Any("output", v)
}

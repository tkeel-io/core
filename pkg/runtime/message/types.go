package message

import (
	"time"

	"github.com/tkeel-io/core/pkg/util"
)

type Message interface {
	String() string
	Promised(interface{})
}

type PromiseFunc = func(interface{})

type MessageBase struct { //nolint
	startTime      time.Time
	PromiseHandler PromiseFunc `json:"-"`
}

func NewBase(promise PromiseFunc) MessageBase {
	return MessageBase{
		startTime:      time.Now(),
		PromiseHandler: promise,
	}
}

func (ms MessageBase) String() string { return "MessageBase" }
func (ms MessageBase) Promised(v interface{}) {
	if nil == ms.PromiseHandler {
		return
	}
	ms.PromiseHandler(v)
}

func (ms MessageBase) Elapsed() *util.ElapsedTime {
	return util.NewElapsedFrom(ms.startTime)
}

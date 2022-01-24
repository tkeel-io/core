package message

import (
	"time"

	"github.com/tkeel-io/core/pkg/util"
)

type Message interface {
	Message()
	Promised(interface{})
}

type PromiseFunc = func(interface{})

type MessageBase struct { //nolint
	startTime      time.Time
	PromiseHandler PromiseFunc `json:"-"`
}

func NewMessageBase(promise PromiseFunc) MessageBase {
	return MessageBase{
		startTime:      time.Now(),
		PromiseHandler: promise,
	}
}

func (ms MessageBase) Message() {}
func (ms MessageBase) Promised(v interface{}) {
	if nil == ms.PromiseHandler {
		return
	}
	ms.PromiseHandler(v)
}

func (ms MessageBase) Elapsed() *util.ElapsedTime {
	return util.NewElapsedFrom(ms.startTime)
}

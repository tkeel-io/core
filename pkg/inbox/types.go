package inbox

import "github.com/tkeel-io/core/pkg/logger"

const (
	defaultNonBlockNum = 10
	defaultExpiredTime = 300 // ms.

	MsgReciverID             = "m-reciverid"
	MsgReciverStatusActive   = "m-active"
	MsgReciverStatusInactive = "m-inactive"
)

var log = logger.NewLogger("core.inbox")

type MessageHandler = func(msg MessageCtx) (int, error)

type Inbox interface {
	Start()
	Stop()
	OnMessage(msg MessageCtx)
}

type Offseter interface {
	Status() bool
	Commit() error
	Confirm()
	AutoCommit() bool
}

type MsgReciver interface {
	Status() string
	OnMessage(msg MessageCtx) (int, error)
}

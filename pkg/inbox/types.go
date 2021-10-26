package inbox

import "github.com/tkeel-io/core/pkg/logger"

const (
	defaultNonBlockNum = 10
	defaultExpiredTime = 300 // ms.
)

var log = logger.NewLogger("core.inbox")

type MessageHandler = func(msg IbElem) (int, error)

type Inbox interface {
	Start()
	Stop()
	OnMessage(msg IbElem)
}

type Offseter interface {
	commit() error
	Status() bool
	Confirm()
}

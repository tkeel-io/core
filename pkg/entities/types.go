package entities

import (
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/statem"
)

var (
	log = logger.NewLogger("core.entities")
)

type EntityOp interface {
	statem.StateMarchiner
}

type EntitySubscriptionOp interface {
	EntityOp

	GetMode() string
}

type WatchKey = mapper.WatchKey

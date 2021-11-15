package entities

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/statem"
)

var (
	log = logger.NewLogger("core.entities")

	errEntityNotFound    = errors.New("entity not found")
	errEmptyEntityMapper = errors.New("empty entity mapper")
)

const (
	MessageCtxHeaderEntityType = "x-entity-type"

	EntityTypeBaseEntity   = "base"
	EntityTypeSubscription = "subscription"
)

type EntityOp interface {
	statem.StateMarchiner
}

type EntitySubscriptionOp interface {
	EntityOp

	GetMode() string
}

type WatchKey = mapper.WatchKey

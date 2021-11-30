package entities

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/statem"
)

var (
	log = logger.NewLogger("core.entities")

	errEntityNotFound      = errors.New("entity not found")
	errEmptyEntityMapper   = errors.New("empty entity mapper")
	errSubscriptionInvalid = errors.New("invalid params")
)

const (
	MessageCtxHeaderEntityType = "x-entity-type"

	EntityTypeBaseEntity   = "base"
	EntityTypeSubscription = "subscription"
)

type EntityOp interface {
	statem.StateMachiner
}

type EntitySubscriptionOp interface {
	EntityOp

	GetMode() string
}

type WatchKey = mapper.WatchKey

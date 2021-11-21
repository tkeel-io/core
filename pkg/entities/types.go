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
	errEntityNotAready   = errors.New("entity not already")
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

// EntityBase statem basic informatinon.
type PropsBase struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Owner   string            `json:"owner"`
	Status  string            `json:"status"`
	Source  string            `json:"source"`
	KValues map[string][]byte `json:"properties"` //nolint
}

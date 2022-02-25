package state

import (
	"sync/atomic"

	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/runtime/environment"
)

type StateContext struct { //nolint
	stateMachine Machiner

	version   int64
	mappers   map[string]mapper.Mapper
	tentacles []mapper.Tentacler
}

func newContext(sm Machiner) StateContext {
	return StateContext{
		stateMachine: sm,
		mappers:      make(map[string]mapper.Mapper),
		tentacles:    make([]mapper.Tentacler, 0),
	}
}

func (ctx *StateContext) LoadEnvironments(env environment.ActorEnv) {
	// load actor mappers.
	ctx.mappers = env.Mappers
	ctx.tentacles = env.Tentacles

	// update version.
	atomic.SwapInt64(&ctx.version, ctx.version+1)
}

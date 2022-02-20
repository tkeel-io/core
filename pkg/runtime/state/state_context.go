package state

import (
	"sync/atomic"

	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/runtime/environment"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type StateContext struct { //nolint
	stateMachine Machiner

	version   int64
	mappers   map[string]mapper.Mapper
	tentacles map[string][]mapper.Tentacler
}

func newContext(sm Machiner) StateContext {
	return StateContext{
		stateMachine: sm,
		mappers:      make(map[string]mapper.Mapper),
		tentacles:    make(map[string][]mapper.Tentacler),
	}
}

func NewContext(sm Machiner, mappers map[string]mapper.Mapper, tentacles []mapper.Tentacler) StateContext {
	stateCtx := StateContext{
		stateMachine: sm,
		mappers:      mappers,
		tentacles:    make(map[string][]mapper.Tentacler),
	}

	for _, tentacle := range tentacles {
		for _, item := range tentacle.Items() {
			stateCtx.tentacles[item.String()] =
				append(stateCtx.tentacles[item.String()], tentacle)
		}
	}

	return stateCtx
}

func (ctx *StateContext) LoadEnvironments(env environment.ActorEnv) {
	ctx.tentacles = make(map[string][]mapper.Tentacler)

	// load actor mappers.
	for _, m := range env.Mappers {
		ctx.mappers[m.ID()] = m
		log.Debug("load environments, mapper ", logger.Eid(ctx.stateMachine.GetID()), zap.String("TQL", m.String()))
	}

	// load actor tentacles.
	for _, t := range env.Tentacles {
		for _, item := range t.Items() {
			ctx.tentacles[item.String()] = append(ctx.tentacles[item.String()], t)
			log.Debug("load environments, watching ", logger.Eid(ctx.stateMachine.GetID()), zap.String("WatchKey", item.String()))
		}
		log.Debug("load environments, tentacle ", logger.Eid(ctx.stateMachine.GetID()), zap.String("tid", t.ID()), zap.String("target", t.TargetID()), zap.String("type", t.Type()), zap.Any("items", t.Items()))
	}

	// update version.
	atomic.SwapInt64(&ctx.version, ctx.version+1)
}

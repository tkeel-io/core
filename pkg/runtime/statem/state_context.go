package statem

import (
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/runtime/environment"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type StateContext struct {
	stateMachine     StateMachiner
	storeClient      IStore
	pubsubClient     IPubsub
	searchClient     ISearch
	timeSeriesClient TSerier
	mappers          map[string]mapper.Mapper      // key=mapperId
	tentacles        map[string][]mapper.Tentacler // key=Sid#propertyKey
}

func NewContext(sm StateMachiner, mappers map[string]mapper.Mapper, tentacles []mapper.Tentacler) StateContext {
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

func (ctx StateContext) StateCliet() IStore {
	return ctx.storeClient
}
func (ctx StateContext) PubsubClient() IPubsub {
	return ctx.pubsubClient
}
func (ctx StateContext) SearchClient() ISearch {
	return ctx.searchClient
}
func (ctx StateContext) TSeriesClient() TSerier {
	return ctx.timeSeriesClient
}

func (ctx StateContext) LoadEnvironments(env environment.ActorEnv) {
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
}

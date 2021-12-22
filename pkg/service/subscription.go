/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/statem"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type SubscriptionService struct {
	pb.UnimplementedSubscriptionServer
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager *entities.EntityManager
}

// NewSubscriptionService returns a new SubscriptionService.
func NewSubscriptionService(ctx context.Context, mgr *entities.EntityManager) (*SubscriptionService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &SubscriptionService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: mgr,
	}, nil
}

func interface2string(in interface{}) (out string) {
	switch inString := in.(type) {
	case string:
		out = inString
	case constraint.Node:
		out = inString.String()
	default:
		out = ""
	}
	return
}

func (s *SubscriptionService) entity2SubscriptionResponse(entity *Entity) (out *pb.SubscriptionResponse) {
	if entity == nil {
		return
	}

	out = &pb.SubscriptionResponse{}

	out.Id = entity.ID
	out.Owner = entity.Owner
	out.Subscription = &pb.SubscriptionObject{}
	out.Subscription.Source = interface2string(entity.KValues[runtime.SubscriptionFieldSource])
	out.Subscription.Filter = interface2string(entity.KValues[runtime.SubscriptionFieldFilter])
	out.Subscription.Target = interface2string(entity.KValues[runtime.SubscriptionFieldTarget])
	out.Subscription.Topic = interface2string(entity.KValues[runtime.SubscriptionFieldTopic])
	out.Subscription.Mode = interface2string(entity.KValues[runtime.SubscriptionFieldMode])
	out.Subscription.PubsubName = interface2string(entity.KValues[runtime.SubscriptionFieldPubsubName])
	return out
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	var entity = new(Entity)

	if req.Id != "" {
		entity.ID = req.Id
	}

	entity.Owner = req.Owner
	entity.Source = req.Source
	entity.Type = runtime.StateMarchineTypeSubscription
	entity.KValues = map[string]constraint.Node{
		runtime.SubscriptionFieldSource:     constraint.StringNode(req.Subscription.Source),
		runtime.SubscriptionFieldFilter:     constraint.StringNode(req.Subscription.Filter),
		runtime.SubscriptionFieldTarget:     constraint.StringNode(req.Subscription.Target),
		runtime.SubscriptionFieldTopic:      constraint.StringNode(req.Subscription.Topic),
		runtime.SubscriptionFieldMode:       constraint.StringNode(req.Subscription.Mode),
		runtime.SubscriptionFieldPubsubName: constraint.StringNode(req.Subscription.PubsubName),
	}

	parseHeaderFrom(ctx, entity)
	if err = s.entityManager.CheckSubscription(ctx, entity); nil != err {
		log.Error("create subscription", zap.Error(err), logger.EntityID(req.Id))
		return
	}

	// set properties.
	if entity, err = s.entityManager.CreateEntity(ctx, entity); nil != err {
		log.Error("create subscription", zap.Error(err), logger.EntityID(req.Id))
		return
	}

	// set mapper.
	entity.Mappers = []statem.MapperDesc{{
		Name:      "subscription",
		TQLString: entity.KValues[runtime.SubscriptionFieldFilter].String(),
	}}

	if _, err = s.entityManager.AppendMapper(ctx, entity); nil != err {
		log.Error("create subscription", zap.Error(err), logger.EntityID(req.Id))
		if _, err0 := s.entityManager.DeleteEntity(ctx, entity); nil != err0 {
			log.Error("destroy subscription", zap.Error(err0), logger.EntityID(req.Id))
		}
		return
	}

	out = s.entity2SubscriptionResponse(entity)

	return out, errors.Wrap(err, "create subscription")
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, req *pb.UpdateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	entity.Type = runtime.StateMarchineTypeSubscription

	entity.KValues = map[string]constraint.Node{
		runtime.SubscriptionFieldSource:     constraint.StringNode(req.Subscription.Source),
		runtime.SubscriptionFieldFilter:     constraint.StringNode(req.Subscription.Filter),
		runtime.SubscriptionFieldTarget:     constraint.StringNode(req.Subscription.Target),
		runtime.SubscriptionFieldTopic:      constraint.StringNode(req.Subscription.Topic),
		runtime.SubscriptionFieldMode:       constraint.StringNode(req.Subscription.Mode),
		runtime.SubscriptionFieldPubsubName: constraint.StringNode(req.Subscription.PubsubName),
	}

	// set properties.
	if entity, err = s.entityManager.SetProperties(ctx, entity); nil != err {
		log.Error("update subscription", zap.Error(err), logger.EntityID(req.Id))
		return
	}

	// set mapper.
	entity.Mappers = []statem.MapperDesc{{
		Name:      "subscription",
		TQLString: entity.KValues[runtime.SubscriptionFieldFilter].String(),
	}}

	if _, err = s.entityManager.AppendMapper(ctx, entity); nil != err {
		log.Error("update subscription", zap.Error(err), logger.EntityID(req.Id))
		return
	}

	out = s.entity2SubscriptionResponse(entity)

	return out, errors.Wrap(err, "update subscription")
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, req *pb.DeleteSubscriptionRequest) (out *pb.DeleteSubscriptionResponse, err error) {
	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	// delete entity.
	if _, err = s.entityManager.DeleteEntity(ctx, entity); nil != err {
		log.Error("delete subscription", zap.Error(err), logger.EntityID(req.Id))
		return
	}

	out = &pb.DeleteSubscriptionResponse{Id: req.Id, Status: "ok"}
	return
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	// delete entity.
	if entity, err = s.entityManager.GetProperties(ctx, entity); nil != err {
		log.Error("get subscription", zap.Error(err), logger.EntityID(req.Id))
		return
	}
	out = s.entity2SubscriptionResponse(entity)
	return
}

func (s *SubscriptionService) ListSubscription(ctx context.Context, req *pb.ListSubscriptionRequest) (out *pb.ListSubscriptionResponse, err error) {
	return &pb.ListSubscriptionResponse{}, nil
}

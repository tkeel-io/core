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
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/util"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const SMTypeSubscription = "SUBSCRIPTION"

type SubscriptionService struct {
	pb.UnimplementedSubscriptionServer
	ctx        context.Context
	cancel     context.CancelFunc
	inited     *atomic.Bool
	apiManager apim.APIManager
}

// NewSubscriptionService returns a new SubscriptionService.
func NewSubscriptionService(ctx context.Context) (*SubscriptionService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &SubscriptionService{
		ctx:    ctx,
		cancel: cancel,
		inited: atomic.NewBool(false),
	}, nil
}

func (s *SubscriptionService) Init(apiManager apim.APIManager) {
	s.apiManager = apiManager
	s.inited.Store(true)
}

func interface2string(in interface{}) (out string) {
	switch inString := in.(type) {
	case string:
		out = inString
	case tdtl.Node:
		out = string(inString.Raw())
	default:
		out = ""
	}
	return
}

func (s *SubscriptionService) entity2SubscriptionResponse(base *apim.BaseRet) (out *pb.SubscriptionResponse) {
	if base == nil {
		return
	}

	out = &pb.SubscriptionResponse{}

	out.Id = base.ID
	out.Owner = base.Owner
	out.Source = base.Source
	out.Subscription = &pb.SubscriptionObject{
		Mode:       interface2string(base.Properties["mode"]),
		Source:     interface2string(base.Properties["source"]),
		Filter:     interface2string(base.Properties["filter"]),
		Target:     interface2string(base.Properties["target"]),
		Topic:      interface2string(base.Properties["topic"]),
		PubsubName: interface2string(base.Properties["pubsub_name"]),
	}
	return out
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	if req.Id == "" {
		req.Id = util.UUID("sub")
	}

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	entity.Type = SMTypeSubscription
	parseHeaderFrom(ctx, entity)
	properties := map[string]interface{}{
		"type":        entity.Type,
		"owner":       entity.Owner,
		"source":      entity.Source,
		"mode":        strings.ToUpper(req.Subscription.Mode),
		"topic":       req.Subscription.Topic,
		"filter":      req.Subscription.Filter,
		"pubsub_name": req.Subscription.PubsubName,
	}

	if entity.Properties, err = json.Marshal(properties); nil != err {
		log.L().Error("create subscription, but invalid params",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
		return out, errors.Wrap(err, "create subscription")
	}

	// set properties.
	var baseRet *apim.BaseRet
	if baseRet, err = s.apiManager.CreateEntity(ctx, entity); nil != err {
		log.L().Error("create subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	mp := &mapper.Mapper{
		ID:          "Subscription",
		TQL:         req.Subscription.Filter,
		Name:        "SubscriptionMapper",
		Owner:       entity.Owner,
		EntityID:    entity.ID,
		Description: "Subscription mapper instance",
	}

	if err = s.apiManager.AppendMapper(ctx, mp); nil != err {
		log.L().Error("create subscription", zap.Error(err), zfield.Eid(req.Id))
		if innerErr := s.apiManager.DeleteEntity(ctx, entity); nil != innerErr {
			log.L().Error("destroy subscription", zap.Error(innerErr), zfield.Eid(req.Id))
		}
		return
	}

	out = s.entity2SubscriptionResponse(baseRet)
	return out, errors.Wrap(err, "create subscription")
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, req *pb.UpdateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	entity.Type = SMTypeSubscription
	parseHeaderFrom(ctx, entity)
	properties := map[string]interface{}{
		"mode":        strings.ToUpper(req.Subscription.Mode),
		"topic":       req.Subscription.Topic,
		"filter":      req.Subscription.Filter,
		"pubsub_name": req.Subscription.PubsubName,
	}

	if entity.Properties, err = json.Marshal(properties); nil != err {
		log.L().Error("create subscription, but invalid params",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
		return out, errors.Wrap(err, "create subscription")
	}

	patches := []*pb.PatchData{{
		Path:     "properties",
		Operator: xjson.OpMerge.String(),
		Value:    entity.Properties,
	}}

	// set properties.
	var baseRet *apim.BaseRet
	if baseRet, _, err = s.apiManager.PatchEntity(ctx, entity, patches); nil != err {
		log.L().Error("update subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	mp := &mapper.Mapper{
		ID:          "Subscription",
		TQL:         req.Subscription.Filter,
		Name:        "SubscriptionMapper",
		Owner:       entity.Owner,
		EntityID:    entity.ID,
		Description: "Subscription mapper instance",
	}

	if err = s.apiManager.AppendMapper(ctx, mp); nil != err {
		log.L().Error("update subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	out = s.entity2SubscriptionResponse(baseRet)
	return out, errors.Wrap(err, "update subscription")
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, req *pb.DeleteSubscriptionRequest) (out *pb.DeleteSubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = SMTypeSubscription
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	if err = s.apiManager.DeleteEntity(ctx, entity); nil != err {
		log.L().Error("delete subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	out = &pb.DeleteSubscriptionResponse{Id: req.Id, Status: "ok"}
	return out, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = SMTypeSubscription
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)

	var baseRet *apim.BaseRet
	if baseRet, err = s.apiManager.GetEntity(ctx, entity); nil != err {
		log.L().Error("get subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}
	out = s.entity2SubscriptionResponse(baseRet)
	return
}

func (s *SubscriptionService) ListSubscription(ctx context.Context, req *pb.ListSubscriptionRequest) (out *pb.ListSubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready")
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	return &pb.ListSubscriptionResponse{}, nil
}

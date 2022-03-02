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
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/runtime"
	"github.com/tkeel-io/core/pkg/runtime/subscription"
	"github.com/tkeel-io/core/pkg/util"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

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
	out.Source = entity.Source
	out.Subscription = &pb.SubscriptionObject{}

	bytes, _ := xjson.EncodeJSON(entity.Properties)
	json.Unmarshal(bytes, &out.Subscription)
	return out
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	if req.Id == "" {
		req.Id = util.UUID("sub")
	}

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	entity.Type = runtime.SMTypeSubscription
	parseHeaderFrom(ctx, entity)
	properties := map[string]interface{}{
		runtime.SMFieldType:                      entity.Type,
		runtime.SMFieldOwner:                     entity.Owner,
		runtime.SMFieldSource:                    entity.Source,
		subscription.SubscriptionFieldMode:       strings.ToUpper(req.Subscription.Mode),
		subscription.SubscriptionFieldTopic:      req.Subscription.Topic,
		subscription.SubscriptionFieldFilter:     req.Subscription.Filter,
		subscription.SubscriptionFieldPubsubName: req.Subscription.PubsubName,
	}

	if entity.Properties, err = parseProps(properties); nil != err {
		log.Error("create subscription, but invalid params",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
		return out, errors.Wrap(err, "create subscription")
	}

	if err = s.apiManager.CheckSubscription(ctx, entity); nil != err {
		log.Error("create subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	// set properties.
	if entity, err = s.apiManager.CreateEntity(ctx, entity); nil != err {
		log.Error("create subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	mp := &dao.Mapper{
		ID:          "Subscription",
		TQL:         req.Subscription.Filter,
		Name:        "SubscriptionMapper",
		Owner:       entity.Owner,
		EntityID:    entity.ID,
		Description: "Subscription mapper instance",
	}

	if err = s.apiManager.AppendMapper(ctx, mp); nil != err {
		log.Error("create subscription", zap.Error(err), zfield.Eid(req.Id))
		if innerErr := s.apiManager.DeleteEntity(ctx, entity); nil != innerErr {
			log.Error("destroy subscription", zap.Error(innerErr), zfield.Eid(req.Id))
		}
		return
	}

	out = s.entity2SubscriptionResponse(entity)
	return out, errors.Wrap(err, "create subscription")
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, req *pb.UpdateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.Source = req.Source
	entity.Type = runtime.SMTypeSubscription
	parseHeaderFrom(ctx, entity)
	properties := map[string]interface{}{
		subscription.SubscriptionFieldFilter:     req.Subscription.Filter,
		subscription.SubscriptionFieldTopic:      req.Subscription.Topic,
		subscription.SubscriptionFieldMode:       strings.ToUpper(req.Subscription.Mode),
		subscription.SubscriptionFieldPubsubName: req.Subscription.PubsubName,
	}

	if entity.Properties, err = parseProps(properties); nil != err {
		log.Error("create subscription, but invalid params",
			zfield.Eid(req.Id), zap.Error(xerrors.ErrInvalidEntityParams))
		return out, errors.Wrap(err, "create subscription")
	}

	// set properties.
	if entity, err = s.apiManager.UpdateEntityProps(ctx, entity); nil != err {
		log.Error("update subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	mp := &dao.Mapper{
		ID:          "Subscription",
		TQL:         req.Subscription.Filter,
		Name:        "SubscriptionMapper",
		Owner:       entity.Owner,
		EntityID:    entity.ID,
		Description: "Subscription mapper instance",
	}

	if err = s.apiManager.AppendMapper(ctx, mp); nil != err {
		log.Error("update subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	out = s.entity2SubscriptionResponse(entity)
	return out, errors.Wrap(err, "update subscription")
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, req *pb.DeleteSubscriptionRequest) (out *pb.DeleteSubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = runtime.SMTypeSubscription
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	if err = s.apiManager.DeleteEntity(ctx, entity); nil != err {
		log.Error("delete subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	// TODO： 不能保证一致性.
	mp := dao.Mapper{
		ID:       "Subscription",
		Owner:    entity.Owner,
		EntityID: entity.ID,
	}

	if err = s.apiManager.RemoveMapper(ctx, &mp); nil != err {
		log.Error("delete subscription, remove mapper", zap.Error(err), zfield.Eid(req.Id))
		return
	}

	out = &pb.DeleteSubscriptionResponse{Id: req.Id, Status: "ok"}
	return out, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready", zfield.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	var entity = new(Entity)
	entity.ID = req.Id
	entity.Type = runtime.SMTypeSubscription
	entity.Owner = req.Owner
	entity.Source = req.Source
	parseHeaderFrom(ctx, entity)
	if entity, err = s.apiManager.GetEntity(ctx, entity); nil != err {
		log.Error("get subscription", zap.Error(err), zfield.Eid(req.Id))
		return
	}
	out = s.entity2SubscriptionResponse(entity)
	return
}

func (s *SubscriptionService) ListSubscription(ctx context.Context, req *pb.ListSubscriptionRequest) (out *pb.ListSubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.Warn("service not ready")
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	return &pb.ListSubscriptionResponse{}, nil
}

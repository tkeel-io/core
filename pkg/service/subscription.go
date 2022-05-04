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
	"fmt"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/core/pkg/logfield"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/util"
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
		log.L().Warn("service not ready", logf.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	log.L().Debug("CreateSubscription", zap.Any("[+]req.Subscription", req.Subscription))
	out = &pb.SubscriptionResponse{}
	if req.Id == "" {
		req.Id = util.UUID("sub")
	}
	sub, err := makeSubscription(req.Subscription)
	if err != nil {
		return out, errors.Wrap(err, "update subscription")
	}
	if sub.ID == "" {
		sub.ID = req.Id
	}
	if sub.Owner == "" {
		sub.Owner = req.Owner
	}

	err = s.apiManager.CreateSubscription(ctx, sub)
	out = &pb.SubscriptionResponse{
		Id:     sub.ID,
		Source: sub.Source,
		Owner:  sub.Owner,
		Subscription: &pb.SubscriptionObject{
			Id: sub.ID,
		},
	}
	log.L().Debug("CreateSubscription", zap.Any("[+]sub", sub), zap.Error(err))
	return out, errors.Wrap(err, "create subscription")
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, req *pb.UpdateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", logf.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	log.L().Debug("UpdateSubscription", zap.Any("[+]req.Subscription", req.Subscription))

	out = &pb.SubscriptionResponse{}
	if req.Id == "" {
		req.Id = util.UUID("sub")
	}
	sub, err := makeSubscription(req.Subscription)
	if err != nil {
		return out, errors.Wrap(err, "update subscription")
	}
	if sub.ID == "" {
		sub.ID = req.Id
	}
	if sub.Owner == "" {
		sub.Owner = req.Owner
	}

	err = s.apiManager.CreateSubscription(ctx, sub)

	log.L().Debug("UpdateSubscription", zap.Any("[+]sub", sub), zap.Error(err))
	return out, errors.Wrap(err, "update subscription")
}

func makeSubscription(subObj *pb.SubscriptionObject) (*repository.Subscription, error) {
	var sub = new(repository.Subscription)
	entitySources, err := entitySources(subObj.Filter)
	if err != nil {
		return nil, errors.Wrap(err, "update subscription")
	}
	if len(entitySources) != 1 {
		return nil, errors.Wrap(err, fmt.Sprintf("subscription source num(%d)!=1", len(entitySources)))
	}

	sub.ID = subObj.Id
	sub.Owner = subObj.Owner
	sub.Source2 = subObj.Source
	sub.Source = subObj.Source
	sub.Mode = subObj.Mode
	sub.Filter = subObj.Filter
	sub.Target = subObj.Target
	sub.Topic = subObj.Topic
	sub.PubsubName = subObj.PubsubName
	for entityID, entityPaths := range entitySources {
		sub.SourceEntityID = entityID
		sub.SourceEntityPaths = entityPaths
		break
	}
	return sub, nil
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, req *pb.DeleteSubscriptionRequest) (out *pb.DeleteSubscriptionResponse, err error) {
	if !s.inited.Load() {
		log.L().Warn("service not ready", logf.Eid(req.Id))
		return nil, errors.Wrap(xerrors.ErrServerNotReady, "service not ready")
	}

	log.L().Debug("DeleteSubscription", zap.Any("[+]req.Subscription", req))
	var sub = new(repository.Subscription)
	if req.Id == "" {
		req.Id = util.UUID("sub")
	}

	sub.ID = req.Id
	sub.Owner = req.Owner
	sub.Source2 = req.Source
	sub, err = s.apiManager.GetSubscription(ctx, sub)
	if err != nil {
		return nil, errors.Wrap(err, "delete subscription")
	}
	err = s.apiManager.DeleteSubscription(ctx, sub)

	log.L().Debug("DeleteSubscription", zap.Any("[+]sub", sub), zap.Error(err))
	out = &pb.DeleteSubscriptionResponse{Id: req.Id, Status: "ok"}
	return out, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	return out, errors.Errorf("Not support GetSubscription")
}

func (s *SubscriptionService) ListSubscription(ctx context.Context, req *pb.ListSubscriptionRequest) (out *pb.ListSubscriptionResponse, err error) {
	return out, errors.Errorf("Not support ListSubscription")
}

func entitySources(filter string) (map[string][]string, error) {
	// cache for node.
	ret, err := tdtl.NewTDTL(filter, nil)
	if err != nil {
		return nil, err
	}
	return ret.Entities(), nil
}

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

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities"
)

type TopicService struct {
	pb.UnimplementedTopicServer
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager *entities.EntityManager
}

const (
	// SubscriptionResponseStatusSuccess means message is processed successfully.
	SubscriptionResponseStatusSuccess = "SUCCESS"
	// SubscriptionResponseStatusRetry means message to be retried by Dapr.
	SubscriptionResponseStatusRetry = "RETRY"
	// SubscriptionResponseStatusDrop means warning is logged and message is dropped.
	SubscriptionResponseStatusDrop = "DROP"
)

func NewTopicService(ctx context.Context, mgr *entities.EntityManager) (*TopicService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &TopicService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: mgr,
	}, nil
}

func (s *TopicService) TopicEventHandler(ctx context.Context, req *pb.TopicEventRequest) (out *pb.TopicEventResponse, err error) {
	var entity = new(Entity)

	entity.KValues = make(map[string]constraint.Node)
	switch kv := req.Data.AsInterface().(type) {
	case map[string]interface{}:
		entity.ID = interface2string(kv["id"])
		entity.Owner = interface2string(kv["owner"])

		switch kvv := kv["data"].(type) {
		case map[string]interface{}:
			for k, v := range kvv {
				entity.KValues[k] = constraint.NewNode(v)
			}
		default:
			return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, nil
		}

	default:
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, nil
	}

	// set properties.
	_, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	return &pb.TopicEventResponse{Status: SubscriptionResponseStatusSuccess}, nil
}

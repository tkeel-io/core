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
	"github.com/tkeel-io/core/pkg/statem"
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
	var values map[string]interface{}
	var properties map[string]constraint.Node
	switch kv := req.Data.AsInterface().(type) {
	case map[string]interface{}:
		values = kv

	default:
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, nil
	}

	// parse data.
	switch data := values["data"].(type) {
	case map[string]interface{}:
		if len(data) > 0 {
			properties = make(map[string]constraint.Node)
			for key, val := range data {
				properties[key] = constraint.NewNode(val)
			}
		}
	default:
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, nil
	}

	msgCtx := statem.MessageContext{
		Headers: statem.Header{},
		Message: statem.PropertyMessage{
			StateID:    interface2string(values["id"]),
			Operator:   constraint.PatchOpReplace.String(),
			Properties: properties,
		},
	}

	msgCtx.Headers.SetTargetID(interface2string(values["id"]))
	msgCtx.Headers.SetOwner(interface2string(values["owner"]))
	msgCtx.Headers.SetOwner(interface2string(values["type"]))
	msgCtx.Headers.SetOwner(interface2string(values["source"]))

	s.entityManager.OnMessage(ctx, msgCtx)
	return &pb.TopicEventResponse{Status: SubscriptionResponseStatusSuccess}, nil
}

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
	"github.com/tkeel-io/core/pkg/entities"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/resource/pubsub/dapr"
	"github.com/tkeel-io/kit/log"
)

type TopicService struct {
	pb.UnimplementedTopicServer
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager entities.EntityManager
}

const (
	// SubscriptionResponseStatusSuccess means message is processed successfully.
	SubscriptionResponseStatusSuccess = "SUCCESS"
	// SubscriptionResponseStatusRetry means message to be retried by Dapr.
	SubscriptionResponseStatusRetry = "RETRY"
	// SubscriptionResponseStatusDrop means warning is logged and message is dropped.
	SubscriptionResponseStatusDrop = "DROP"
)

func NewTopicService(ctx context.Context, entityManager entities.EntityManager) (*TopicService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &TopicService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: entityManager,
	}, nil
}

func (s *TopicService) TopicEventHandler(ctx context.Context, req *pb.TopicEventRequest) (out *pb.TopicEventResponse, err error) {
	log.Debug("catched event", zfield.ReqID(req.Meta.Id), zfield.Type(req.Meta.Type),
		zfield.Pubsub(req.Meta.Pubsubname), zfield.Topic(req.Meta.Topic), zfield.Source(req.Meta.Source))
	res, err := dapr.HandleEvent(ctx, req)
	return res, errors.Wrap(err, "handle event")
}

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
	"github.com/tkeel-io/collectjs"
	pb "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/resource/pubsub/dapr"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type TopicService struct {
	pb.UnimplementedTopicServer
	ctx        context.Context
	cancel     context.CancelFunc
	apiManager apim.APIManager
}

const (
	// SubscriptionResponseStatusSuccess means message is processed successfully.
	SubscriptionResponseStatusSuccess = "SUCCESS"
	// SubscriptionResponseStatusRetry means message to be retried by Dapr.
	SubscriptionResponseStatusRetry = "RETRY"
	// SubscriptionResponseStatusDrop means warning is logged and message is dropped.
	SubscriptionResponseStatusDrop = "DROP"
)

func NewTopicService(ctx context.Context) (*TopicService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &TopicService{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (s *TopicService) Init(apiManager apim.APIManager) {
	s.apiManager = apiManager
}

func (s *TopicService) TopicEventHandler(ctx context.Context, req *pb.TopicEventRequest) (out *pb.TopicEventResponse, err error) {
	panic("never used")
}

func (s *TopicService) TopicClusterEventHandler(ctx context.Context, req *pb.TopicEventRequest) (out *pb.TopicEventResponse, err error) {
	log.L().Debug("received event", zfield.ReqID(req.Meta.Id),
		zfield.Type(req.Meta.Type), zfield.Source(req.Meta.Source),
		zfield.Topic(req.Meta.Topic), zfield.Pubsub(req.Meta.Pubsubname))

	id, _, _ := collectjs.Get(req.RawData, "id")
	typ, _, _ := collectjs.Get(req.RawData, "type")
	owner, _, _ := collectjs.Get(req.RawData, "owner")
	source, _, _ := collectjs.Get(req.RawData, "source")

	var payload []byte
	// set event payload.
	if payload, _, err = collectjs.Get(req.RawData, "data.rawData"); nil != err {
		log.L().Warn("get event payload", zap.String("id", req.Meta.Id), zap.Any("event", req), zfield.Reason(err.Error()))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, errors.Wrap(err, "get event payload")
	}

	var ev pb.ProtoEvent
	ev.SetType(pb.ETEntity)
	ev.SetAttr(pb.MetaTopic, req.Meta.Topic)
	ev.SetAttr(pb.MetaEntityID, string(id))
	ev.SetAttr(pb.MetaOwner, string(owner))
	ev.SetAttr(pb.MetaSource, string(source))
	ev.SetAttr(pb.MetaEntityType, string(typ))
	ev.SetPayload(&pb.ProtoEvent_Patches{
		Patches: &pb.PatchDatas{
			Patches: []*pb.PatchData{{
				Path:     "properties.rawData",
				Operator: "replace",
				Value:    payload,
			}},
		},
	})

	res, err := dapr.HandleEvent(ctx, &ev)
	if nil != err {
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, errors.Wrap(err, "handle event")
	}

	return res, nil
}

type RawData struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Mark      string `json:"mark"`
	Path      string `json:"path"`
	Values    string `json:"values"`
	Timestamp int64  `json:"ts"` //nolint
}

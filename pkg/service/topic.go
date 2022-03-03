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

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	"github.com/tkeel-io/collectjs"
	pb "github.com/tkeel-io/core/api/core/v1"
	zfield "github.com/tkeel-io/core/pkg/logger"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource/pubsub/dapr"
	"github.com/tkeel-io/core/pkg/runtime/message"
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
	// parse CloudEvent from pb.TopicEventRequest.
	log.Debug("received event", zfield.ReqID(req.Meta.Id), zfield.Type(req.Meta.Type),
		zfield.Pubsub(req.Meta.Pubsubname), zfield.Topic(req.Meta.Topic), zfield.Source(req.Meta.Source))

	ev := cloudevents.NewEvent()
	err = ev.UnmarshalJSON(req.RawData)
	if nil != err {
		log.Warn("data must be CloudEvents spec", zap.String("id", req.Meta.Id), zap.Any("event", req), zfield.Reason(err.Error()))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, errors.Wrap(err, "unmarshal event")
	}

	ev.SetExtension(message.ExtCloudEventTopic, req.Meta.Topic)
	ev.SetExtension(message.ExtCloudEventPubsub, req.Meta.Pubsubname)
	ev.SetExtension(message.ExtCloudEventConsumerType, dao.ConsumerTypeCore.String())
	res, err := dapr.Get().DeliveredEvent(ctx, ev)
	return res, errors.Wrap(err, "handle event")
}

func (s *TopicService) TopicClusterEventHandler(ctx context.Context, req *pb.TopicEventRequest) (out *pb.TopicEventResponse, err error) {
	log.Debug("received event", zfield.ReqID(req.Meta.Id),
		zfield.Type(req.Meta.Type), zfield.Source(req.Meta.Source),
		zfield.Topic(req.Meta.Topic), zfield.Pubsub(req.Meta.Pubsubname))

	m := make(map[string]interface{})
	if err = json.Unmarshal(req.RawData, &m); nil != err {
		log.Warn("unmarshal data", zap.String("id", req.Meta.Id), zap.Any("event", req), zfield.Reason(err.Error()))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, errors.Wrap(err, "unmarshal data")
	}

	ev := cloudevents.NewEvent()
	ev.SetID(req.Meta.Id)
	ev.SetType(message.MessageTypeRaw.String())
	ev.SetSource(req.Meta.Source)

	// set extension fields.
	ev.SetExtension(message.ExtEntityID, m["id"])
	ev.SetExtension(message.ExtEntityType, m["type"])
	ev.SetExtension(message.ExtEntityOwner, m["owner"])
	ev.SetExtension(message.ExtEntitySource, m["source"])
	ev.SetExtension(message.ExtCloudEventConsumerType, dao.ConsumerTypeDispatch.String())

	var payload []byte
	// set event payload.
	if payload, _, err = collectjs.Get(req.RawData, "data.rawData"); nil != err {
		log.Warn("get event payload", zap.String("id", req.Meta.Id), zap.Any("event", req), zfield.Reason(err.Error()))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, errors.Wrap(err, "get event payload")
	}

	ev.SetDataContentType(cloudevents.ApplicationCloudEventsJSON)
	if err = ev.SetData(payload); nil != err {
		log.Warn("set event payload", zap.String("id", req.Meta.Id), zap.Any("event", req), zfield.Reason(err.Error()))
		return &pb.TopicEventResponse{Status: SubscriptionResponseStatusDrop}, errors.Wrap(err, "set event payload")
	}

	res, err := dapr.Get().DeliveredEvent(ctx, ev)
	return res, errors.Wrap(err, "handle event")
}

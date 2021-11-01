package service

import (
	"context"
	"strings"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/entities"
)

// TopicEventService is a dapr pubsub subscription service.
type TopicEventService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager *entities.EntityManager
}

// NewTopicEventService returns a new TopicEventService.
func NewTopicEventService(ctx context.Context, mgr *entities.EntityManager) (*TopicEventService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &TopicEventService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: mgr,
	}, nil
}

// Name return the name.
func (e *TopicEventService) Name() string {
	return "topic-event"
}

// RegisterService register some methods.
func (e *TopicEventService) RegisterService(daprService common.Service) (err error) {
	// register all handlers.
	if err = e.AddSubTopic(daprService, "core-pub", "core-pubsub"); nil != err {
		return
	}
	return
}

func (e *TopicEventService) AddSubTopic(daprService common.Service, topic, pubsubName string) (err error) {
	sub := &common.Subscription{
		PubsubName: pubsubName,
		Topic:      topic,
		Route:      "/",
		Metadata:   map[string]string{},
	}
	if err = daprService.AddTopicEventHandler(sub, e.topicHandler); err != nil {
		return
	}
	return
}

func getSourceFrom(pubsubName string) (source string) {
	return strings.Split(pubsubName, "-")[0]
}

func TopicEvent2EntityContext(in *common.TopicEvent) (out *entities.EntityContext, err error) {
	ec := entities.NewEntityContext(nil)
	var entityID, owner, plugin string

	log.Infof("dispose event, pubsub: %s. topic: %s, datatype: %T, data: %v.",
		in.PubsubName, in.Topic, in.Data, in.Data)

	if in.DataContentType == "application/json" {
		inData, ok := in.Data.(map[string]interface{})
		if !ok {
			return nil, errTypeError
		}

		// get entity id.
		switch entityIds := inData["entity_id"].(type) {
		case string:
			entityID = entityIds
		default:
			return nil, errTypeError
		}

		// get entity owner.
		switch tempOwner := inData["owner"].(type) {
		case string:
			owner = tempOwner
		default:
			err = errTypeError
			return
		}

		// get entity source plugin.
		plugin = getSourceFrom(in.PubsubName)

		/*
			switch tempPlugin := inData["plugin"].(type) {
			case string:
				plugin = tempPlugin
			default:
				err = errTypeError
				return
			}
		*/

		// get entity data.
		switch tempData := inData["data"].(type) {
		case string, []byte:
			values := make(map[string]interface{})
			values["__data__"] = tempData
			ec.Message = &entities.EntityMessage{SourceID: entityID, Values: values}
		case map[string]interface{}:
			ec.Message = &entities.EntityMessage{SourceID: entityID, Values: tempData}
		default:
			err = errTypeError
			return
		}

		ec.Headers.SetOwner(owner)
		ec.Headers.SetPluginID(plugin)
		ec.Headers.SetTargetID(entityID)
	}
	return &ec, nil
}

func (e *TopicEventService) topicHandler(ctx context.Context, in *common.TopicEvent) (retry bool, err error) {
	if ec, err := TopicEvent2EntityContext(in); err != nil {
		return false, err
	} else if in.DataContentType == "application/json" {
		e.entityManager.SendMsg(*ec)
	}

	return false, nil
}

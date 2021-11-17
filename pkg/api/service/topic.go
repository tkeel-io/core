package service

import (
	"context"
	"strings"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/constraint"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/statem"
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

func TopicEvent2EntityContext(in *common.TopicEvent) (out *statem.MessageContext, err error) {
	ec := statem.MessageContext{}
	var entityID, owner string

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
		_ = getSourceFrom(in.PubsubName)

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

			// 这里对2进制编码是存在问题的.
			ec.Message = statem.NewPropertyMessage(entityID, tempConvert(values))
		case map[string]interface{}:
			// 临时处理为了先完成entity-config的功能.

			ec.Message = statem.NewPropertyMessage(entityID, tempConvert(tempData))
		default:
			err = errTypeError
			return
		}

		ec.Headers = statem.Header{}
		ec.Headers.SetOwner(owner)
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

func tempConvert(values map[string]interface{}) map[string][]byte {
	ret := make(map[string][]byte)
	for key, val := range values {
		ret[key] = []byte(constraint.NewNode(val).String())
	}
	return ret
}

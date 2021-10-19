package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/service"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

const (
	entityFieldID          = "id"
	entityFieldUserID      = "user_id"
	entityFieldSource      = "source"
	entityFieldTag         = "tag"
	entityFieldVersion     = "version"
	entityStateFieldPrefix = "__internal_"
)

func internalFieldName(fieldName string) string {
	return fmt.Sprintf("%s%s", entityStateFieldPrefix, fieldName)
}

type EntityServiceConfig struct {
	TableName   string
	StateName   string
	BindingName string
}

// EntityService is a time-series service.
type EntityService struct {
	daprClient   dapr.Client
	tableName    string
	stateName    string
	bindingName  string
	entityManger *entities.EntityManager
}

// NewEntityService returns a new EntityService.
func NewEntityService(entityConfig *EntityServiceConfig, manager *entities.EntityManager) (*EntityService, error) {
	cli, err := dapr.NewClient()
	if err != nil {
		err = errors.Wrap(err, "start dapr clint err")
	}

	return &EntityService{
		daprClient:   cli,
		tableName:    entityConfig.TableName,
		stateName:    entityConfig.StateName,
		bindingName:  entityConfig.BindingName,
		entityManger: manager,
	}, err
}

// Name return the name.
func (e *EntityService) Name() string {
	return "entity"
}

// RegisterService register some methods.
func (e *EntityService) RegisterService(daprService common.Service) (err error) {
	// register all handlers.
	if err = daprService.AddServiceInvocationHandler("/plugins/{plugin}/entities/{entity}", e.entityHandler); nil != err {
		return
	}
	if err = daprService.AddServiceInvocationHandler("/plugins/{plugin}/entities", e.entitiesHandler); nil != err {
		return
	}
	if err = e.AddSubTopic(daprService, "core", "core-pubsub"); nil != err {
		return
	}
	return
}

func (e *EntityService) AddSubTopic(daprService common.Service, topic, pubsubName string) (err error) {
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

func TopicEvent2EntityContext(in *common.TopicEvent) (out *entities.EntityContext, err error) {
	ec := entities.EntityContext{}
	var entityID, userID string
	ec.Headers = make(map[string]string)
	if in.DataContentType == "application/json" {
		inData, ok := in.Data.(map[string]interface{})
		if !ok {
			return nil, errTypeError
		}
		switch entityIds := inData["entity_id"].(type) {
		case string:
			entityID = entityIds
		default:
			return nil, errTypeError
		}
		switch tempUserID := inData["tenant_id"].(type) {
		case string:
			userID = tempUserID
		default:
			err = errTypeError
			return
		}
		switch tempData := inData["data"].(type) {
		case string, []byte:
			values := make(map[string]interface{})
			values["__data__"] = tempData
			ec.Message = &entities.EntityMsg{SourceID: "", Values: values}
		case map[string]interface{}:
			ec.Message = &entities.EntityMsg{SourceID: "", Values: tempData}
		default:
			err = errTypeError
			return
		}

		ec.Headers["user_id"] = userID
		ec.SetTarget(entityID)
	}
	return &ec, nil
}

func (e *EntityService) topicHandler(ctx context.Context, in *common.TopicEvent) (retry bool, err error) {
	if ec, err := TopicEvent2EntityContext(in); err != nil {
		return false, err
	} else if in.DataContentType == "application/json" {
		e.entityManger.SendMsg(*ec)
	}

	return false, nil
}

// Echo test for RegisterService.
func (e *EntityService) entityHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}

	log.Info("call entity handler.", in.Verb, in.QueryString, in.DataTypeURL, string(in.Data))

	switch in.Verb {
	case http.MethodGet:
		return e.entityGet(ctx, in)
	case http.MethodPost:
		return e.entityCreate(ctx, in)
	case http.MethodPut:
		return e.entityUpdate(ctx, in)
	case http.MethodDelete:
		return e.entityDelete(ctx, in)
		// temporary.
	case http.MethodPatch:
		return e.entityCreate(ctx, in)
	default:
	}

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func invocationEvent2Entity(ctx context.Context, in *common.InvocationEvent) (entity *entities.EntityBase, err error) {
	items, _ := url.ParseQuery(in.QueryString)
	entity = &entities.EntityBase{}
	var ok bool
	if ctxValue := ctx.Value(service.ContextKey(service.Entity)); ctx != nil {
		if entity.ID, ok = ctxValue.(string); !ok {
			err = errors.New("type error")
		}
	}

	if ctxValue := ctx.Value(service.ContextKey(service.Plugin)); ctx != nil {
		if entity.Source, ok = ctxValue.(string); !ok {
			err = errors.New("type error")
		}
	}
	if len(items[service.User]) > 0 {
		entity.UserID = items[service.User][0]
	}
	return
}

// EntityGet returns an entity information.
func (e *EntityService) entityGet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		entity *entities.EntityBase
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	if entity, err = invocationEvent2Entity(ctx, in); nil != err {
		return
	}
	e2, err := e.entityManger.GetAllProperties(ctx, entity.ID)
	if err != nil {
	} else {
		out.Data, _ = json.Marshal(e2)
	}

	return
}

// EntityGet create  an entity.
func (e *EntityService) entityCreate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		entity  entities.EntityBase
		kvalues = make(map[string]interface{})
	)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	// TODO 创建entity

	kvalues[internalFieldName(entityFieldTag)] = entity.Tag
	kvalues[internalFieldName(entityFieldID)] = entity.ID
	kvalues[internalFieldName(entityFieldUserID)] = entity.UserID
	kvalues[internalFieldName(entityFieldSource)] = entity.Source
	kvalues[internalFieldName(entityFieldVersion)] = entity.Version

	// encode kvs.
	if out.Data, err = json.Marshal(kvalues); nil != err {
		return
	}

	return
}

// entityUpdate update an entity.
func (e *EntityService) entityUpdate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	return out, errors.Wrap(err, "json marshall failed")
}

// EntityGet delete an entity.
func (e *EntityService) entityDelete(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)
	return out, errors.Wrap(err, "dapr client delete state failed")
}

// Echo test for RegisterService.
func (e *EntityService) entitiesHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return out, err
	}

	// parse request query...

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return out, err
}

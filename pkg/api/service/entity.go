package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/tkeel-io/core/utils"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

const (
	entityFieldID     = "id"
	entityFieldTag    = "tag"
	entityFieldType   = "type"
	entityFieldUserID = "user_id"
	entityFieldSource = "source"
)

var errBodyMustBeJSON = errors.New("request body must be json")

func entityFieldRequired(fieldName string) error {
	return fmt.Errorf("entity field(%s) required", fieldName)
}

type Entity = entities.EntityBase

// EntityService is a time-series service.
type EntityService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	daprClient    dapr.Client
	entityManager *entities.EntityManager
}

// NewEntityService returns a new EntityService.
func NewEntityService(ctx context.Context, mgr *entities.EntityManager) (*EntityService, error) {
	ctx, cancel := context.WithCancel(ctx)

	cli, err := dapr.NewClient()
	if err != nil {
		err = errors.Wrap(err, "start dapr clint err")
	}

	return &EntityService{
		ctx:           ctx,
		cancel:        cancel,
		daprClient:    cli,
		entityManager: mgr,
	}, err
}

// Name return the name.
func (e *EntityService) Name() string {
	return "entity"
}

// RegisterService register some methods.
func (e *EntityService) RegisterService(daprService common.Service) error {
	// register all handlers.
	if err := daprService.AddServiceInvocationHandler("entities", e.entityHandler); nil != err {
		return errors.Wrap(err, "dapr add 'entities' service invocation handler err")
	}
	if err := daprService.AddServiceInvocationHandler("entitylist", e.entityList); nil != err {
		return errors.Wrap(err, "dapr add 'entitylist' service invocation handler err")
	}
	return nil
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

func (e *EntityService) getValFromValues(values url.Values, key string) (string, error) {
	if vals, exists := values[key]; exists && len(vals) > 0 {
		return vals[0], nil
	}

	return "", entityFieldRequired(key)
}

func getStringFrom(ctx context.Context, key string) (string, error) {
	if val := ctx.Value(service.ContextKey(key)); nil != val {
		if v, ok := val.(string); ok {
			return v, nil
		}
	}
	return "", entityFieldRequired(key)
}

func (e *EntityService) getEntityFrom(ctx context.Context, entity *Entity, in *common.InvocationEvent, idRequired bool) error { // nolint
	var (
		err    error
		values url.Values
	)

	if values, err = url.ParseQuery(in.QueryString); nil != err {
		return errors.Wrap(err, "parse URL failed")
	}

	if entity.Type, err = getStringFrom(ctx, service.HeaderType); nil == err {
		// type field required.
		log.Info("parse http request field(type) from header successes.")
	} else if entity.Type, err = e.getValFromValues(values, entityFieldType); nil != err {
		log.Error("parse http request field(type) from query failed", ctx, err)
		return err
	}

	if entity.Source, err = getStringFrom(ctx, service.HeaderSource); nil == err {
		// source field required.
		log.Info("parse http request field(source) from header successes.")
	} else if entity.Source, err = e.getValFromValues(values, entityFieldSource); nil != err {
		log.Error("parse http request field(source) from query failed", ctx, err)
		return err
	}

	if entity.UserID, err = getStringFrom(ctx, service.HeaderUser); nil == err {
		// userId field required.
		log.Info("parse http request field(user) from header successed.")
	} else if entity.UserID, err = e.getValFromValues(values, entityFieldUserID); nil != err {
		log.Error("parse http request field(user) from query failed", ctx, err)
		return err
	}

	if entity.ID, err = e.getValFromValues(values, entityFieldID); nil != err {
		// entity id field
		if !idRequired {
			entity.ID = utils.GenerateUUID()
		}
	}

	// tags
	if vals, exists := values[entityFieldTag]; exists && len(vals) > 0 {
		tag := strings.Join(vals, ";")
		entity.Tag = &tag
	}

	return err
}

// EntityGet returns an entity information.
func (e *EntityService) entityGet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	err = e.getEntityFrom(ctx, entity, in, true)
	if nil != err {
		return
	}

	// get entity from entity manager.
	entity, err = e.entityManager.GetAllProperties(ctx, entity)
	if nil != err {
		log.Errorf("get entity failed, %s", err.Error())
		return
	}

	// encode entity.
	out.Data, err = json.Marshal(entity)

	return
}

// EntityGet create  an entity.
func (e *EntityService) entityCreate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	err = e.getEntityFrom(ctx, entity, in, false)
	if nil != err {
		return
	}

	if len(in.Data) > 0 {
		entity.KValues = make(map[string]interface{})
		if err = json.Unmarshal(in.Data, &entity.KValues); nil != err {
			return out, errBodyMustBeJSON
		}
	}

	// set properties.
	entity, err = e.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	// encode kvs.
	out.Data, err = json.Marshal(entity)

	return out, errors.Wrap(err, "entity create failed")
}

// entityUpdate update an entity.
func (e *EntityService) entityUpdate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	err = e.getEntityFrom(ctx, entity, in, false)
	if nil != err {
		return
	}

	if len(in.Data) > 0 {
		entity.KValues = make(map[string]interface{})
		if err = json.Unmarshal(in.Data, &entity.KValues); nil != err {
			return out, errBodyMustBeJSON
		}
	}

	// set properties.
	entity, err = e.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	// encode kvs.
	out.Data, err = json.Marshal(entity)

	return out, errors.Wrap(err, "entity update failed")
}

// EntityGet delete an entity.
func (e *EntityService) entityDelete(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	err = e.getEntityFrom(ctx, entity, in, false)
	if nil != err {
		return
	}

	// delete entity.
	entity, err = e.entityManager.DeleteEntity(ctx, entity)
	if nil != err {
		return
	}

	// encode kvs.
	out.Data, err = json.Marshal(entity)

	return
}

// Echo test for RegisterService.
func (e *EntityService) entityList(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
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

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/service"

	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

// EntityService is a time-series service.
type EntityService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager *entities.EntityManager
}

// NewEntityService returns a new EntityService.
func NewEntityService(ctx context.Context, mgr *entities.EntityManager) (*EntityService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &EntityService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: mgr,
	}, nil
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
	} else if err = daprService.AddServiceInvocationHandler("/plugins/{plugin}/entities", e.entitiesHandler); nil != err {
		return
	} else if err = daprService.AddServiceInvocationHandler("/plugins/{plugin}/entities/{entity}/mappers", e.AppendMapper); nil != err {
		return
	}

	return
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

func (e *EntityService) getEntityFrom(ctx context.Context, entity *Entity, in *common.InvocationEvent, idRequired bool) (source string, err error) { // nolint
	var values url.Values

	if values, err = url.ParseQuery(in.QueryString); nil != err {
		return source, errors.Wrap(err, "parse URL failed")
	}

	if entity.Type, err = getStringFrom(ctx, service.HeaderType); nil == err {
		// type field required.
		log.Info("parse http request field(type) from header successes.")
	} else if entity.Type, err = e.getValFromValues(values, entityFieldType); nil != err {
		log.Error("parse http request field(type) from query failed", values, ctx, err)
		return source, err
	}

	if entity.PluginID, err = getStringFrom(ctx, service.Plugin); nil != err {
		// plugin field required.
		log.Error("parse http request field(source) from path failed", ctx, err)
		return source, err
	}

	if entity.Owner, err = getStringFrom(ctx, service.HeaderOwner); nil == err {
		// owner field required.
		log.Info("parse http request field(owner) from header successed.")
	} else if entity.Owner, err = e.getValFromValues(values, entityFieldOwner); nil != err {
		log.Error("parse http request field(owner) from query failed", ctx, err)
		return source, err
	}

	if source, err = getStringFrom(ctx, service.HeaderSource); nil == err {
		// source field required.
		log.Info("parse http request field(source) from header successed.")
	} else if source, err = e.getValFromValues(values, entityFieldSource); nil != err {
		log.Error("parse http request field(source) from query failed", ctx, err)
		return source, err
	}

	if entity.ID, err = getStringFrom(ctx, service.Entity); nil == err {
		log.Info("parse http request field(id) from path successed.")
	} else if entity.ID, err = e.getValFromValues(values, entityFieldID); nil != err {
		if !idRequired {
			err = nil
		} else {
			log.Error("parse http request field(id) from query failed", ctx, err)
		}
	}

	return source, err
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

	_, err = e.getEntityFrom(ctx, entity, in, true)
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

	_, err = e.getEntityFrom(ctx, entity, in, false)
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

	_, err = e.getEntityFrom(ctx, entity, in, true)
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

	_, err = e.getEntityFrom(ctx, entity, in, true)
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

// EntityList List entities.
func (e *EntityService) entityList(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	// TODO

	defer errResult(out, err)
	return
}

func (e *EntityService) AppendMapper(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	_, err = e.getEntityFrom(ctx, entity, in, false)
	if nil != err {
		return
	}

	if len(in.Data) > 0 {
		mapperDesc := entities.MapperDesc{}
		if err = json.Unmarshal(in.Data, &mapperDesc); nil != err {
			return out, errBodyMustBeJSON
		}
		entity.Mappers = []entities.MapperDesc{mapperDesc}
	}

	// set properties.
	entity, err = e.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	// encode kvs.
	out.Data, err = json.Marshal(entity)

	return out, errors.Wrap(err, "append mapper failed")
}

// Echo test for RegisterService.
func (e *EntityService) entitiesHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return out, err
	}

	switch in.Verb {
	case http.MethodPost:
		return e.entityCreate(ctx, in)
	case http.MethodGet:
		return e.entityList(ctx, in)
	default:
	}

	// parse request query...

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return out, err
}

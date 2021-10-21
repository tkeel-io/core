package service

import (
	"context"
	"net/http"
	"net/url"

	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/service"

	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

// SubscriptionService is a subscription manage service.
type SubscriptionService struct {
	ctx           context.Context
	cancel        context.CancelFunc
	entityManager *entities.EntityManager
}

// NewSubscriptionService returns a new SubscriptionService.
func NewSubscriptionService(ctx context.Context, mgr *entities.EntityManager) (*SubscriptionService, error) {
	ctx, cancel := context.WithCancel(ctx)

	return &SubscriptionService{
		ctx:           ctx,
		cancel:        cancel,
		entityManager: mgr,
	}, nil
}

// Name return the name.
func (s *SubscriptionService) Name() string {
	return "subscription"
}

// RegisterService register some methods.
func (s *SubscriptionService) RegisterService(daprService common.Service) (err error) {
	// register all handlers.
	if err = daprService.AddServiceInvocationHandler("/plugins/{plugin}/subscriptions/{entity}", s.subscriptionHandler); nil != err {
		return
	}
	if err = daprService.AddServiceInvocationHandler("/plugins/{plugin}/subscriptions", s.subscriptionsHandler); nil != err {
		return
	}
	return
}

// Echo test for RegisterService.
func (s *SubscriptionService) subscriptionHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return
	}

	log.Info("call entity handler.", in.Verb, in.QueryString, in.DataTypeURL, string(in.Data))

	switch in.Verb {
	case http.MethodGet:
		return s.subscriptionGet(ctx, in)
	case http.MethodPut:
		return s.subscriptionUpdate(ctx, in)
	case http.MethodDelete:
		return s.subscriptionDelete(ctx, in)
	default:
	}

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func (s *SubscriptionService) getValFromValues(values url.Values, key string) (string, error) {
	if vals, exists := values[key]; exists && len(vals) > 0 {
		return vals[0], nil
	}

	return "", entityFieldRequired(key)
}

func (s *SubscriptionService) getEntityFrom(ctx context.Context, entity *Entity, in *common.InvocationEvent, idRequired bool) (source string, err error) { // nolint
	var values url.Values

	if values, err = url.ParseQuery(in.QueryString); nil != err {
		return source, errors.Wrap(err, "parse URL failed")
	}

	if entity.Type, err = getStringFrom(ctx, service.HeaderType); nil == err {
		// type field required.
		log.Info("parse http request field(type) from header successes.")
	} else if entity.Type, err = s.getValFromValues(values, entityFieldType); nil != err {
		log.Error("parse http request field(type) from query failed", values, ctx, err)
		return source, err
	}

	if entity.PluginID, err = getStringFrom(ctx, service.Plugin); nil != err {
		// plugin field required.
		log.Error("parse http request field(source) from path failed", ctx, err)
		return source, err
	}

	if entity.Owner, err = getStringFrom(ctx, service.HeaderUser); nil == err {
		// userId field required.
		log.Info("parse http request field(owner) from header successed.")
	} else if entity.Owner, err = s.getValFromValues(values, entityFieldOwner); nil != err {
		log.Error("parse http request field(owner) from query failed", ctx, err)
		return source, err
	}

	if source, err = getStringFrom(ctx, service.HeaderSource); nil == err {
		// userId field required.
		log.Info("parse http request field(source) from header successed.")
	} else if source, err = s.getValFromValues(values, entityFieldSource); nil != err {
		log.Error("parse http request field(source) from query failed", ctx, err)
		return source, err
	}

	if entity.ID, err = getStringFrom(ctx, service.Entity); nil == err {
		log.Info("parse http request field(id) from path successed.")
	} else if entity.ID, err = s.getValFromValues(values, entityFieldID); nil != err {
		if !idRequired {
			err = nil
		} else {
			log.Error("parse http request field(id) from query failed", ctx, err)
		}
	}

	return source, err
}

// EntityGet returns an entity information.
func (s *SubscriptionService) subscriptionGet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	_, err = s.getEntityFrom(ctx, entity, in, true)
	if nil != err {
		return
	}

	return
}

// EntityGet create  an entity.
func (s *SubscriptionService) subscriptionCreate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	_, err = s.getEntityFrom(ctx, entity, in, false)
	if nil != err {
		return
	}

	return out, errors.Wrap(err, "entity create failed")
}

// subscriptionUpdate update an entity.
func (s *SubscriptionService) subscriptionUpdate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	_, err = s.getEntityFrom(ctx, entity, in, true)
	if nil != err {
		return
	}

	return out, errors.Wrap(err, "entity update failed")
}

// EntityGet delete an entity.
func (s *SubscriptionService) subscriptionDelete(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var entity = new(Entity)

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	defer errResult(out, err)

	_, err = s.getEntityFrom(ctx, entity, in, true)
	if nil != err {
		return
	}

	return
}

// EntityList List entities.
func (s *SubscriptionService) subscriptionList(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	// TODO

	defer errResult(out, err)
	return
}

// Echo test for RegisterService.
func (s *SubscriptionService) subscriptionsHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("nil invocation parameter")
		return out, err
	}

	switch in.Verb {
	case http.MethodPost:
		return s.subscriptionCreate(ctx, in)
	case http.MethodGet:
		return s.subscriptionList(ctx, in)
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

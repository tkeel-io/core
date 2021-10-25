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

func (s *SubscriptionService) getEntityFrom(ctx context.Context, entity *Entity, in *common.InvocationEvent, idRequired bool) (source string, err error) {
	var values url.Values

	if values, err = url.ParseQuery(in.QueryString); nil != err {
		return source, errors.Wrap(err, "parse URL failed")
	}

	if entity.PluginID, err = getStringFrom(ctx, service.Plugin); nil != err {
		// plugin field required.
		log.Error("parse http request field(pluginId) from path failed", ctx, err)
		return source, err
	}

	if entity.Owner, err = getStringFrom(ctx, service.HeaderOwner); nil == err {
		// owner field required.
		log.Info("parse http request field(owner) from header successed.")
	} else if entity.Owner, err = s.getValFromValues(values, entityFieldOwner); nil != err {
		log.Error("parse http request field(owner) from query failed", ctx, err)
		return source, err
	}

	if source, err = getStringFrom(ctx, service.HeaderSource); nil == err {
		// source field required.
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

	_, err = s.getEntityFrom(ctx, entity, in, false)
	if nil != err {
		return
	}

	// get entity from entity manager.
	entity, err = s.entityManager.GetAllProperties(ctx, entity)
	if nil != err {
		log.Errorf("get entity failed, %s", err.Error())
		return
	}

	// encode entity.
	if out.Data, err = json.Marshal(entity); nil != err {
		log.Errorf("create subscription failed, %s.", err.Error())
		return out, errors.Wrap(err, "create subscription failed")
	}

	return
}

// EntityGet create  an entity.
func (s *SubscriptionService) subscriptionCreate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		entity       = new(Entity)
		subscription *entities.SubscriptionBase
	)

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

	subscription, err = DecodeSubscription(in.Data)
	if nil != err {
		log.Errorf("invalid request, %s.", err.Error())
		return out, errors.Wrap(err, "invalid request")
	}

	entity.Type = entities.EntityTypeSubscription
	entity.KValues = map[string]interface{}{
		entities.SubscriptionFieldSource:     subscription.Source,
		entities.SubscriptionFieldFilter:     subscription.Filter,
		entities.SubscriptionFieldTarget:     subscription.Target,
		entities.SubscriptionFieldTopic:      subscription.Topic,
		entities.SubscriptionFieldMode:       subscription.Mode,
		entities.SubscriptionFieldPubsubName: subscription.PubsubName,
	}

	// set properties.
	entity, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	// encode kvs.
	if out.Data, err = json.Marshal(entity); nil != err {
		log.Errorf("create subscription failed, %s.", err.Error())
		return out, errors.Wrap(err, "create subscription failed")
	}

	return out, nil
}

// subscriptionUpdate update an entity.
func (s *SubscriptionService) subscriptionUpdate(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		entity       = new(Entity)
		subscription *entities.SubscriptionBase
	)

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

	subscription, err = DecodeSubscription(in.Data)
	if nil != err {
		log.Errorf("invalid request, %s.", err.Error())
		return out, errors.Wrap(err, "invalid request")
	}

	entity.Type = entities.EntityTypeSubscription
	entity.KValues = map[string]interface{}{
		"source": subscription.Source,
		"filter": subscription.Filter,
		"target": subscription.Target,
		"mode":   subscription.Mode,
	}

	// set properties.
	entity, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	// encode kvs.
	if out.Data, err = json.Marshal(entity); nil != err {
		log.Errorf("update subscription failed, %s.", err.Error())
		return out, errors.Wrap(err, "update subscription failed")
	}

	return out, nil
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

	// delete entity.
	entity, err = s.entityManager.DeleteEntity(ctx, entity)
	if nil != err {
		return
	}

	// encode kvs.
	out.Data, err = json.Marshal(entity)

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

func DecodeSubscription(data []byte) (*entities.SubscriptionBase, error) {
	subscription := entities.SubscriptionBase{}
	err := json.Unmarshal(data, &subscription)
	return &subscription, errors.Wrap(err, "decode subscription base information failed, request body must be json")
}

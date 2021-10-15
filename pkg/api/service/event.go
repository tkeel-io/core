package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/tkeel-io/core/pkg/model"
	"github.com/tkeel-io/core/pkg/service"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	"github.com/pkg/errors"
)

type EventServiceConfig struct {
	RawTopic          string
	TimeSeriesTopic   string
	PropertyTopic     string
	RelationShipTopic string
	StoreName         string
	PubsubName        string
}

type EventService struct {
	cli               dapr.Client
	rawTopic          string
	tsTopic           string
	propertyTopic     string
	relationShipTopic string
	storeName         string
	pubsubName        string
}

func NewEventService(conf *EventServiceConfig) (*EventService, error) {
	cli, err := dapr.NewClient()

	return &EventService{
		cli:               cli,
		rawTopic:          conf.RawTopic,
		tsTopic:           conf.TimeSeriesTopic,
		propertyTopic:     conf.PropertyTopic,
		relationShipTopic: conf.RelationShipTopic,
		storeName:         conf.StoreName,
		pubsubName:        conf.PubsubName,
	}, errors.Unwrap(err)
}

// Name return the name.
func (s *EventService) Name() string {
	return "event"
}

// RegisterService register some method.
func (s *EventService) RegisterService(daprService common.Service) error {
	// register all handlers.
	if err := daprService.AddServiceInvocationHandler("event", s.eventHandler); nil != err {
		return errors.Wrap(err, "dapr service in vocation handler err")
	}
	return nil
}

func (s *EventService) eventHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	if in == nil {
		err = errors.New("invocation parameter required")
		return
	}

	switch in.Verb {
	case http.MethodGet:
		return s.getEvent(ctx, in)
	case http.MethodPost:
		return s.writeEvent(ctx, in)
	}

	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

func (s *EventService) getEvent(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	var (
		ok                  bool
		topic, source, user string
	)

	if topic, ok = ctx.Value(service.HeaderTopic).(string); !ok || topic == "" {
		return nil, model.ErrTopicNil
	}

	if source, ok = ctx.Value(service.HeaderSource).(string); !ok || source == "" {
		return nil, model.ErrSourceNil
	}

	if user, ok = ctx.Value(service.HeaderUser).(string); !ok || user == "" {
		return nil, model.ErrUserNil
	}

	data, err := s.cli.GetState(context.Background(), s.storeName, source+user+topic)
	if err != nil {
		log.Error(err)
	}
	out = &common.Content{
		Data:        data.Value,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}

//nolint:cyclop
func (s *EventService) writeEvent(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	ev, err := model.NewKEventFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "new KEvent failed")
	}

	query, err := url.ParseQuery(in.QueryString)
	if err != nil {
		return nil, errors.Wrap(err, "query parse err")
	}
	ev.Type = query.Get(service.QueryType)
	ev.Data = in.Data

	var (
		source, topic, user string
		ok                  bool
	)
	if source, ok = ctx.Value(service.HeaderSource).(string); !ok {
		log.Errorf("ctx: '%' parse to string err", service.HeaderSource)
	}

	if topic, ok = ctx.Value(service.HeaderTopic).(string); !ok {
		log.Errorf("ctx: '%' parse to string err", service.HeaderTopic)
	}

	if user, ok = ctx.Value(service.HeaderUser).(string); !ok {
		log.Errorf("ctx: '%' parse to string err", service.HeaderUser)
	}

	err = s.cli.SaveState(context.Background(), s.storeName, source+user+topic, in.Data)
	if err != nil {
		log.Error(err)
	}

	var pubsubTopic string
	data, err := json.Marshal(ev)
	if err != nil {
		log.Error(err)
	}
	switch ev.Type {
	case model.EventTypeRaw:
		pubsubTopic = s.rawTopic
	case model.EventTypeProperty:
		pubsubTopic = s.propertyTopic
	case model.EventTypeTS:
		pubsubTopic = s.tsTopic
	case model.EventTypeRelationship:
		pubsubTopic = s.relationShipTopic
	default:
		return nil, model.ErrEventType
	}

	err = s.cli.PublishEvent(context.Background(), s.pubsubName, pubsubTopic, data)
	if err != nil {
		log.Error(err)
	}
	return out, errors.Wrap(err, "client publish err")
}

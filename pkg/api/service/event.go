package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	dapr "github.com/dapr/go-sdk/client"

	"github.com/tkeel-io/core/pkg/model"
	"github.com/tkeel-io/core/pkg/service"
	"github.com/dapr/go-sdk/service/common"
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
	}, err
}

// Name return the name.
func (this *EventService) Name() string {
	return "event"
}

// RegisterService register some method
func (this *EventService) RegisterService(daprService common.Service) error {
	//register all handlers.
	if err := daprService.AddServiceInvocationHandler("event", this.eventHandler); nil != err {
		return err
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
	topic := ctx.Value(service.HeaderTopic).(string)
	if topic == "" {
		return nil, model.TopicNilErr
	}

	source := ctx.Value(service.HeaderSource).(string)
	if source == "" {
		return nil, model.SourceNilErr
	}

	user := ctx.Value(service.HeaderUser).(string)
	if user == "" {
		return nil, model.UserNilErr
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

func (s *EventService) writeEvent(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	ev, err := model.NewKEventFromContext(ctx)
	if err != nil {
		return nil, err
	}

	query, err := url.ParseQuery(in.QueryString)
	if err != nil {
		return nil, err
	}
	ev.Type = query.Get(service.QueryType)
	ev.Data = in.Data

	source := ctx.Value(service.HeaderSource).(string)
	topic := ctx.Value(service.HeaderTopic).(string)
	user := ctx.Value(service.HeaderUser).(string)

	err = s.cli.SaveState(context.Background(), s.storeName, source+user+topic, in.Data)
	if err != nil {
		log.Error(err)
	}

	pubsubTopic := ""
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
		return nil, model.EventTypeErr

	}

	err = s.cli.PublishEvent(context.Background(), s.pubsubName, pubsubTopic, data)
	if err != nil {
		log.Error(err)
	}
	return
}

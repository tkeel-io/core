package service

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/runtime/message"
)

type ProxyService struct {
	pb.UnimplementedProxyServer
	entityManager entities.EntityManager
}

func NewProxyService(entityManager entities.EntityManager) *ProxyService {
	return &ProxyService{entityManager: entityManager}
}

func (p *ProxyService) Route(ctx context.Context, in *pb.RouteRequest) (*pb.RouteResponse, error) {
	err := p.entityManager.OnMessage(ctx, constructEvent(in))
	return &pb.RouteResponse{}, errors.Wrap(err, "route message")
}

func constructEvent(in *pb.RouteRequest) cloudevents.Event {
	ev := cloudevents.NewEvent()
	ev.SetID(in.Header[message.ExtCloudEventID])
	ev.SetSpecVersion(in.Header[message.ExtCloudEventSpec])
	ev.SetType(in.Header[message.ExtCloudEventType])
	ev.SetSource(in.Header[message.ExtCloudEventSource])
	ev.SetSubject(in.Header[message.ExtCloudEventSubject])
	ev.SetDataSchema(in.Header[message.ExtCloudEventDataSchema])
	ev.SetDataContentType(in.Header[message.ExtCloudEventContentType])

	ev.SetData(in.Data)

	return ev
}

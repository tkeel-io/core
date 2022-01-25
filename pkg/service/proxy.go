package service

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/entities"
)

type ProxyService struct {
	pb.UnimplementedProxyServer
	entityManager entities.EntityManager
}

func NewProxyService(entityManager entities.EntityManager) *ProxyService {
	return &ProxyService{entityManager: entityManager}
}

func (p *ProxyService) Route(ctx context.Context, in *pb.RouteRequest) (*pb.RouteResponse, error) {
	ev := cloudevents.NewEvent()
	err := p.entityManager.OnMessage(ctx, ev)
	return &pb.RouteResponse{}, errors.Wrap(err, "route message")
}

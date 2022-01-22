package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/entities"
	"github.com/tkeel-io/core/pkg/runtime/statem"
)

type ProxyService struct {
	pb.UnimplementedProxyServer
	entityManager entities.EntityManager
}

func NewProxyService(entityManager entities.EntityManager) *ProxyService {
	return &ProxyService{entityManager: entityManager}
}

func (p *ProxyService) Route(ctx context.Context, in *pb.RouteRequest) (*pb.RouteResponse, error) {
	msgCtx := statem.MessageContext{}
	err := p.entityManager.OnMessage(ctx, msgCtx)
	return &pb.RouteResponse{}, errors.Wrap(err, "route message")
}

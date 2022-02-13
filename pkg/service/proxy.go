package service

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/manager/holder"
)

type ProxyService struct {
	pb.UnimplementedProxyServer
	apiManager apim.APIManager
}

func NewProxyService(apiManager apim.APIManager) *ProxyService {
	return &ProxyService{apiManager: apiManager}
}

func (p *ProxyService) Respond(ctx context.Context, in *pb.RespondRequest) (*pb.RespondResponse, error) {
	p.apiManager.OnRespond(ctx, &holder.Response{
		Metadata: in.Metadata,
		Data:     in.Data,
	})

	return &pb.RespondResponse{}, nil
}

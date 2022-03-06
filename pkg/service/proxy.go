package service

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	apim "github.com/tkeel-io/core/pkg/manager"
	"github.com/tkeel-io/core/pkg/manager/holder"
	"github.com/tkeel-io/core/pkg/types"
)

type ProxyService struct {
	pb.UnimplementedProxyServer
	apiManager apim.APIManager
}

func NewProxyService() *ProxyService {
	return &ProxyService{}
}

func (p *ProxyService) Init(apiManager apim.APIManager) {
	p.apiManager = apiManager
}

func (p *ProxyService) Respond(ctx context.Context, in *pb.RespondRequest) (*pb.RespondResponse, error) {
	reqID := in.Metadata[pb.MetaRequestID]
	status := in.Metadata[pb.MetaResponseStatus]
	errCode := in.Metadata[pb.MetaResponseErrCode]

	p.apiManager.OnRespond(ctx, &holder.Response{
		ID:       reqID,
		Status:   types.Status(status),
		ErrCode:  errCode,
		Metadata: in.Metadata,
		Data:     in.Data,
	})

	return &pb.RespondResponse{}, nil
}

package service

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
)

type EntityService struct {
	pb.UnimplementedEntityServer
}

func NewEntityService() *EntityService {
	return &EntityService{}
}

func (s *EntityService) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.CreateEntityResponse, error) {
	return &pb.CreateEntityResponse{}, nil
}
func (s *EntityService) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.UpdateEntityResponse, error) {
	return &pb.UpdateEntityResponse{}, nil
}
func (s *EntityService) DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (*pb.DeleteEntityResponse, error) {
	return &pb.DeleteEntityResponse{}, nil
}
func (s *EntityService) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (*pb.GetEntityResponse, error) {
	return &pb.GetEntityResponse{}, nil
}
func (s *EntityService) ListEntity(ctx context.Context, req *pb.ListEntityRequest) (*pb.ListEntityResponse, error) {
	return &pb.ListEntityResponse{}, nil
}

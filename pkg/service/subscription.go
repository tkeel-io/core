package service

import (
	"context"

	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/entities"
)

type SubscriptionService struct {
	pb.UnimplementedSubscriptionServer
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

func interface2string(in interface{}) (out string) {
	switch inString := in.(type) {
	case string:
		out = inString
	default:
		out = ""
	}
	return
}

func (s *SubscriptionService) entity2SubscriptionResponse(entity *Entity) (out *pb.SubscriptionResponse) {
	if entity == nil {
		return
	}

	out = &pb.SubscriptionResponse{}

	out.Plugin = entity.PluginID
	out.Owner = entity.Owner
	out.Id = entity.ID
	out.Subscription = &pb.SubscriptionObject{}
	out.Subscription.Source = interface2string(entity.KValues[entities.SubscriptionFieldSource])
	out.Subscription.Filter = interface2string(entity.KValues[entities.SubscriptionFieldFilter])
	out.Subscription.Target = interface2string(entity.KValues[entities.SubscriptionFieldTarget])
	out.Subscription.Topic = interface2string(entity.KValues[entities.SubscriptionFieldTopic])
	out.Subscription.Mode = interface2string(entity.KValues[entities.SubscriptionFieldMode])
	out.Subscription.PubsubName = interface2string(entity.KValues[entities.SubscriptionFieldPubsubName])
	return out
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req *pb.CreateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	var entity = new(Entity)

	if req.Id != "" {
		entity.ID = req.Id
	}
	entity.Owner = req.Owner
	entity.PluginID = req.Plugin
	entity.Type = entities.EntityTypeSubscription

	entity.KValues = map[string]interface{}{
		entities.SubscriptionFieldSource:     req.Subscription.Source,
		entities.SubscriptionFieldFilter:     req.Subscription.Filter,
		entities.SubscriptionFieldTarget:     req.Subscription.Target,
		entities.SubscriptionFieldTopic:      req.Subscription.Topic,
		entities.SubscriptionFieldMode:       req.Subscription.Mode,
		entities.SubscriptionFieldPubsubName: req.Subscription.PubsubName,
	}

	// set properties.
	entity, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	out = s.entity2SubscriptionResponse(entity)

	return
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, req *pb.UpdateSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.PluginID = req.Plugin
	entity.Type = entities.EntityTypeSubscription

	entity.KValues = map[string]interface{}{
		entities.SubscriptionFieldSource:     req.Subscription.Source,
		entities.SubscriptionFieldFilter:     req.Subscription.Filter,
		entities.SubscriptionFieldTarget:     req.Subscription.Target,
		entities.SubscriptionFieldTopic:      req.Subscription.Topic,
		entities.SubscriptionFieldMode:       req.Subscription.Mode,
		entities.SubscriptionFieldPubsubName: req.Subscription.PubsubName,
	}

	// set properties.
	entity, err = s.entityManager.SetProperties(ctx, entity)
	if nil != err {
		return
	}

	out = s.entity2SubscriptionResponse(entity)

	return
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, req *pb.DeleteSubscriptionRequest) (out *pb.DeleteSubscriptionResponse, err error) {
	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.PluginID = req.Plugin
	// delete entity.
	_, err = s.entityManager.DeleteEntity(ctx, entity)
	if nil != err {
		return
	}

	out = &pb.DeleteSubscriptionResponse{Id: req.Id, Status: "ok"}
	return
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (out *pb.SubscriptionResponse, err error) {
	var entity = new(Entity)

	entity.ID = req.Id
	entity.Owner = req.Owner
	entity.PluginID = req.Plugin
	// delete entity.
	entity, err = s.entityManager.GetAllProperties(ctx, entity)
	if nil != err {
		return
	}
	out = s.entity2SubscriptionResponse(entity)
	return
}

func (s *SubscriptionService) ListSubscription(ctx context.Context, req *pb.ListSubscriptionRequest) (out *pb.ListSubscriptionResponse, err error) {
	return &pb.ListSubscriptionResponse{}, nil
}

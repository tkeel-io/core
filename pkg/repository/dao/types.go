package dao

import (
	"context"

	"go.etcd.io/etcd/api/v3/mvccpb"
)

const EntityStorePrefix = "CORE.ENTITY"

var (
	EntityTypeBasic        = "BASIC"
	EntityTypeSubscription = "SUBSCRIPTION"

	PUT    EnventType = EnventType(mvccpb.PUT)
	DELETE EnventType = EnventType(mvccpb.DELETE)
)

type EnventType mvccpb.Event_EventType

func (et EnventType) String() string {
	return mvccpb.Event_EventType(et).String()
}

type ListMapperReq struct {
	Owner    string
	EntityID string
}

type IDao interface {
	Get(ctx context.Context, id string) (en *Entity, err error)
	Put(ctx context.Context, en *Entity) error
	Exists(ctx context.Context, id string) error
}

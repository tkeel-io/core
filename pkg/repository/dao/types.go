package dao

import (
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

type IDao interface {
}

type DecodeFunc func([]byte) (Resource, error)
type RangeResourceFunc func([]*mvccpb.KeyValue)
type WatchResourceFunc func(EnventType, *mvccpb.KeyValue)

type Resource interface {
	Codec() KVCodec
}

package runtime

import (
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	proto "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func deliveredEvent(msg *sarama.ConsumerMessage) (proto.Event, error) {
	var ev proto.ProtoEvent
	if err := proto.Unmarshal(msg.Value, &ev); nil != err {
		log.Error("decode Event", zap.Error(err))
		return nil, errors.Wrap(err, "decode event")
	}

	return &ev, nil
}

type EventType string

const (
	ETCache        EventType = "core.event.Cache"
	ETEntity       EventType = "core.event.Entity"
	ETRuntime      EventType = "core.event.Runtime"
	ETMapperCreate EventType = "core.event.Mapper.Create"
	ETMapperUpdate EventType = "core.event.Mapper.Update"
	ETMapperDelete EventType = "core.event.Mapper.Delete"
	ETEntityCreate EventType = "core.event.Entity.Create"
	ETEntityUpdate EventType = "core.event.Entity.Update"
	ETEntityDelete EventType = "core.event.Entity.Delete"
)

type PatchOp string

const (
	OpUndef   PatchOp = "undefine"
	OpAdd     PatchOp = "add"
	OpTest    PatchOp = "test"
	OpCopy    PatchOp = "copy"
	OpMove    PatchOp = "move"
	OpMerge   PatchOp = "merge"
	OpRemove  PatchOp = "remove"
	OpReplace PatchOp = "replace"
)

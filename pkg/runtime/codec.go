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

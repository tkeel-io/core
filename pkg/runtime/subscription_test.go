package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/repository/dao"
	xjson "github.com/tkeel-io/core/pkg/util/json"
	"github.com/tkeel-io/tdtl"
)

func Test_makePayload(t *testing.T) {
	// "id":           ev.Attr(v1.MetaSender),
	// "subscribe_id": ev.Entity(),
	// "type":         ev.Attr(v1.MetaEntityType),
	// "owner":        ev.Attr(v1.MetaOwner),
	// "source":       ev.Attr(v1.MetaSource),
	ev := &v1.ProtoEvent{
		Metadata: map[string]string{
			v1.MetaEntityID:   "sub123",
			v1.MetaOwner:      "admin",
			v1.MetaSource:     "core",
			v1.MetaSender:     "device123",
			v1.MetaEntityType: dao.EntityTypeSubscription,
		},
	}

	bytes, err := makePayload(ev, "sub123", []Patch{
		{
			Op:    xjson.OpMerge,
			Path:  "properties.temps",
			Value: tdtl.New(`{"temp":20}`),
		},
		{
			Op:    xjson.OpReplace,
			Path:  "properties.metrics.cpu.value",
			Value: tdtl.New(`0.78`),
		},
	})

	assert.Nil(t, err)
	t.Log("payload: ", string(bytes))
}

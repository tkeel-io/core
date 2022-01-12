package statem

import (
	"testing"
	"time"

	msgpack "github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/constraint"
)

func TestEncodeBase(t *testing.T) {
	vals := map[string]interface{}{
		"int":    uint8(123),
		"string": "string",
		"object": map[interface{}]interface{}{"a": "v"},
		"binary": []byte("1234567"),
	}

	bytes, _ := msgpack.Marshal(vals)

	var v map[string]interface{}
	msgpack.Unmarshal(bytes, &v)
	assert.Equal(t, vals, v)
}

func TestDecodeBase(t *testing.T) {
	base := &Base{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Mappers:  []MapperDesc{{Name: "test", TQLString: "insert into device123 select device234.temp as temp"}},
		KValues: map[string]constraint.Node{
			"temp": constraint.NewNode(123.3),
		},
		Configs: make(map[string]constraint.Config),
	}

	bytes, _ := EncodeBase(base)
	out, _ := DecodeBase(bytes)
	assert.Equal(t, base, out)
}

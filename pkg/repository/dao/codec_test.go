package dao

import (
	"encoding/json"
	"testing"
	"time"

	msgpack "github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/constraint"
)

func TestMsgPack(t *testing.T) {
	vals := map[string]interface{}{
		"int":    uint8(123),
		"string": "string",
		"object": map[interface{}]interface{}{"a": "v"},
		"binary": []byte("1234567"),
	}

	// different json: json can't encode map[interface{}]interface{}.
	bytes, _ := msgpack.Marshal(vals)

	var v map[string]interface{}
	msgpack.Unmarshal(bytes, &v)
	assert.Equal(t, vals, v)
}

func TestEncode(t *testing.T) {
	enc := entityCodec{}

	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]constraint.Node{
			"temp": constraint.NewNode(123.3),
		},
	}

	bytes, err := enc.Encode(&en)
	assert.Nil(t, err, "error must be nil")
	assert.NotNil(t, bytes, "result can not nil")
	t.Log("encode result: ", string(bytes))
}

func TestDecode(t *testing.T) {
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]constraint.Node{
			"temp": constraint.NewNode(123.3),
		},
	}

	enc := entityCodec{}
	bytes, err := enc.Encode(&en)
	assert.Nil(t, err, "error must be nil")

	var res Entity
	err = enc.Decode(bytes, &res)
	assert.Nil(t, err, "error must be nil")
	assert.Equal(t, en, res)
}

func BenchmarkEncode(b *testing.B) {
	enc := entityCodec{}
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]constraint.Node{
			"temp": constraint.NewNode(123.3),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Encode(&en)
	}
}

func BenchmarkDecode(b *testing.B) {
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]constraint.Node{
			"temp": constraint.NewNode(123.3),
			"metrics0": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics1": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics2": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics3": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics4": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics5": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics6": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics7": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
		},
	}

	enc := entityCodec{}
	bytes, err := enc.Encode(&en)
	assert.Nil(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Decode(bytes, &en)
	}
}

func BenchmarkJsonEncode(b *testing.B) {
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]constraint.Node{
			"temp": constraint.NewNode(123.3),
			"metrics0": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics1": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics2": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics3": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics4": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics5": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics6": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
			"metrics7": constraint.NewNode(map[string]interface{}{
				"cpu_used": 0.2,
				"mem_used": 0.7,
			}),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(en)
	}
}

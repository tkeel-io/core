package dao

import (
	"encoding/json"
	"testing"
	"time"

	msgpack "github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tdtl"
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
		Properties: map[string]tdtl.Node{
			"temp": tdtl.FloatNode(123.3),
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
		Properties: map[string]tdtl.Node{
			"temp": tdtl.FloatNode(123.3),
		},
	}

	enc := entityCodec{}
	bytes, err := enc.Encode(&en)
	assert.Nil(t, err, "error must be nil")

	var res Entity
	err = enc.Decode(bytes, &res)
	assert.Nil(t, err, "error must be nil")
	t.Log("decode result: ", res.Properties)
}

func BenchmarkEncode(b *testing.B) {
	enc := entityCodec{}
	en := Entity{
		ID:       "device123",
		Type:     "BASIC",
		Owner:    "admin",
		Source:   "dm",
		LastTime: time.Now().UnixNano() / 1e6,
		Properties: map[string]tdtl.Node{
			"temp":     tdtl.FloatNode(12345.77777),
			"metrics0": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics1": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics2": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics3": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics4": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics5": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics6": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics7": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
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
		Properties: map[string]tdtl.Node{
			"temp":     tdtl.FloatNode(12345.77777),
			"metrics0": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics1": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics2": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics3": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics4": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics5": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics6": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics7": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
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
		Properties: map[string]tdtl.Node{
			"temp":     tdtl.FloatNode(12345.77777),
			"metrics0": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics1": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics2": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics3": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics4": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics5": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics6": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
			"metrics7": tdtl.New(`{"cpu_used": 0.2, "mem_used": 0.78}`),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(en)
	}
}

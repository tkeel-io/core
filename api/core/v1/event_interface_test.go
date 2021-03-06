// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: api/core/v1/event.proto

package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarshal1(t *testing.T) {
	ev := &ProtoEvent{
		Id:        "ev-12345",
		Timestamp: time.Now().UnixNano(),
		Metadata:  map[string]string{},
		Data: &ProtoEvent_RawData{
			RawData: []byte(`{"name": "tomas"}`),
		},
	}

	bytes, err := Marshal(ev)
	assert.Nil(t, err)

	// unmarshal .
	var e ProtoEvent
	err = Unmarshal(bytes, &e)
	assert.Nil(t, err)
	assert.Equal(t, "ev-12345", e.Id)
	assert.Equal(t, []byte(`{"name": "tomas"}`), e.GetRawData())
}

func TestMarshal2(t *testing.T) {
	ev := &ProtoEvent{
		Id:        "ev-12345",
		Timestamp: time.Now().UnixNano(),
		Metadata:  map[string]string{},
		Data: &ProtoEvent_Patches{
			Patches: &PatchDatas{
				Patches: []*PatchData{
					{
						Path:     "metrics.temp",
						Operator: "replace",
						Value:    []byte(`20`),
					},
				},
			},
		},
	}

	bytes, err := Marshal(ev)
	assert.Nil(t, err)

	// unmarshal .
	var e ProtoEvent
	err = Unmarshal(bytes, &e)
	assert.Nil(t, err)
	assert.Equal(t, "ev-12345", e.Id)

	patches := e.GetPatches().GetPatches()

	assert.Equal(t, 1, len(patches))
	assert.Equal(t, "metrics.temp", patches[0].Path)
}

func TestMarshal3(t *testing.T) {
	ev := &ProtoEvent{
		Id:        "ev-12345",
		Timestamp: time.Now().UnixNano(),
		Metadata:  map[string]string{},
		Data: &ProtoEvent_SystemData{
			SystemData: &SystemData{
				Operator: "Create",
				Data:     []byte(`{"id": "en-123", "properties": {"temp": 20}}`),
			},
		},
	}

	bytes, err := Marshal(ev)
	assert.Nil(t, err)

	// unmarshal .
	var e ProtoEvent
	err = Unmarshal(bytes, &e)
	assert.Nil(t, err)
	assert.Equal(t, "ev-12345", e.Id)

	t.Log(string(e.Action().GetData()))

}

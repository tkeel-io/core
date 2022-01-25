package message

import (
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	ev := cloudevents.NewEvent()
	ev.SetID("test-event123")
	ev.SetSource("test")
	ev.SetType("test")
	ev.SetData(map[string]interface{}{
		"metrics": map[string]interface{}{
			"cpu_used": 0.3,
			"mem_used": 0.4,
		},
	})

	bytes, err := ev.MarshalJSON()
	assert.Nil(t, err)
	t.Log(string(bytes))
}

func TestEventMarshalJSON(t *testing.T) {
	ev := cloudevents.NewEvent()
	ev.SetID("test-event123")
	ev.SetSource("test")
	ev.SetType("test")
	ev.SetExtension("extenid", "123")
	ev.SetExtension("extentype", "DEVICE")
	ev.SetExtension("extowner", "admin")
	ev.SetExtension("extsource", "core")
	ev.SetExtension("extmsgid", "msg-xxxxx")
	ev.SetData([]byte(`{"a":"b"}`))
	bytes, _ := ev.MarshalJSON()

	e := cloudevents.NewEvent()
	err := e.UnmarshalJSON(bytes)
	assert.Nil(t, err)
	t.Log(string(bytes))
	t.Log(e)
}

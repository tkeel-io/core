package message

import (
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestEvent(t *testing.T) {
	ev := cloudevents.NewEvent()
	ev.SetID("test-event123")
	ev.SetData(cloudevents.EncodingBinary.String(), map[string]interface{}{
		"metrics": map[string]interface{}{
			"cpu_used": 0.3,
			"mem_used": 0.4,
		},
	})

	bytes, _ := ev.MarshalJSON()

	t.Log(string(bytes))
}

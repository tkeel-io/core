package dapr

// import (
// 	"context"
// 	"log"
// 	"testing"

// 	cloudevents "github.com/cloudevents/sdk-go"
// 	daprSDK "github.com/dapr/go-sdk/client"
// )

// func TestSend(t *testing.T) {
// 	// create dapr client.
// 	daprClient, err := daprSDK.NewClient()
// 	if nil != err {
// 		log.Fatal(err)
// 	}
// 	// create an event.
// 	ev := cloudevents.NewEvent()

// 	ev.SetID("uuid-123")
// 	ev.SetType("publish")
// 	ev.SetSource("device-manager")
// 	ev.SetDataContentType(cloudevents.ApplicationJSON)
// 	ev.SetData(map[string]interface{}{
// 		"id":     "device123",
// 		"type":   "DEVICE",
// 		"owner":  "admin",
// 		"source": "device-manager",
// 		"temp":   20,
// 	})

// 	// check event.
// 	if nil != ev.Validate() {
// 		log.Fatal(err)
// 	}

// 	// set some information for component.
// 	metadata := make(map[string]string)

// 	daprClient.PublishEvent(
// 		context.Background(),
// 		"pubsubName", "topicName", ev,
// 		daprSDK.PublishEventWithMetadata(metadata),
// 		daprSDK.PublishEventWithContentType(cloudevents.ApplicationJSON))
// }

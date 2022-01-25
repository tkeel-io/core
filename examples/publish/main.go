package main

import (
	"context"
	"fmt"

	dapr "github.com/dapr/go-sdk/client"
)

func main() {
	client, _ := dapr.NewClient()
	var data = map[string]interface{}{
		"id":     "device123",
		"owner":  "admin",
		"type":   "BASICE",
		"source": "devices",
		"connectinfo": map[string]interface{}{
			"temp": "123",
		},
	}

	err := client.PublishEvent(context.Background(), "core-pubsub", "core-pub", data)
	fmt.Println(err)

}

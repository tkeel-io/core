package main

import (
	"context"
	"fmt"
	"time"

	dapr "github.com/dapr/go-sdk/client"
)

type ValueType struct {
	Value interface{} `json:"value"`
	Time  int64       `json:"time"`
}

func main() {
	client, err := dapr.NewClient()
	if nil != err {
		panic(err)
	}

	// create entity.
	createUrl := "plugins/pluginA/entities?id=test1&owner=abc&source=abc&type=device"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		createUrl,
		"POST",
		&dapr.DataContent{
			ContentType: "application/json",
		})
	if nil != err {
		panic(err)
	}
	fmt.Println(string(result))

	// get entity.
	getUrl := "plugins/pluginA/entities/test1?owner=abc&source=abc&type=device"

	result, err = client.InvokeMethodWithContent(context.Background(),
		"core",
		getUrl,
		"GET",
		&dapr.DataContent{
			ContentType: "application/json",
		})
	if nil != err {
		panic(err)
	}
	fmt.Println(string(result))

	// pub entity.

	data := make(map[string]interface{})
	data["entity_id"] = "test1"
	data["owner"] = "abc"
	dataItem := make(map[string]interface{})
	dataItem["core"] = ValueType{Value: 189, Time: time.Now().UnixNano() / 1e6}
	data["data"] = dataItem

	err = client.PublishEvent(context.Background(),
		"client-pubsub",
		"core-pub",
		data,
	)

	if nil != err {
		panic(err)
	}

	// get entity.
	result, err = client.InvokeMethodWithContent(context.Background(),
		"core",
		getUrl,
		"GET",
		&dapr.DataContent{
			ContentType: "application/json",
		})
	if nil != err {
		panic(err)
	}
	fmt.Println(string(result))
}

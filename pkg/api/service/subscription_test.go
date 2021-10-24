package service

import (
	"context"
	"fmt"
	"testing"

	dapr "github.com/dapr/go-sdk/client"
)

func TestSubscriptionCreate(t *testing.T) {
	queryString := "source=pubsubm&owner=admin"
	methodName := "/plugins/pubsubm/subscriptions/sub123"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		fmt.Sprintf("%s?%s", methodName, queryString),
		"POST",
		&dapr.DataContent{
			ContentType: "application/json",
			Data: []byte(`{
				"source": "device-management",
				"filter": "select * where thing_id=abcd",
				"target": "pubsubm",
				"mode": "realtime"
			}`),
		})

	if nil != err {
		t.Log("output: ", err)
	}

	t.Log("output: ", string(result))
}

func TestSubscriptionUpdate(t *testing.T) {
	queryString := "source=pubsubm&owner=admin"
	methodName := "/plugins/pubsubm/subscriptions/sub123"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		fmt.Sprintf("%s?%s", methodName, queryString),
		"PUT",
		&dapr.DataContent{
			ContentType: "application/json",
			Data: []byte(`{
				"filter": "select *",
			}`),
		})

	if nil != err {
		t.Log("output: ", err)
	}

	t.Log("output: ", string(result))
}

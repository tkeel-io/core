package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	dapr "github.com/dapr/go-sdk/client"
)

var client dapr.Client

func TestMain(m *testing.M) {
	var err error
	client, err = dapr.NewClient()
	if nil != err {
		panic(err)
	}

	m.Run()

	os.Exit(0)
}

func TestEntityCreate(t *testing.T) {
	queryString := "id=test123&source=abcd&type=DEVICE&user_id=admin"
	methodName := "entities"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		fmt.Sprintf("%s?%s", methodName, queryString),
		"POST",
		&dapr.DataContent{
			ContentType: "application/json",
			Data: []byte(`{
				"name": "tempSensor1",
				"owner": "tomas"
			}`),
		})

	if nil != err {
		t.Log("output: ", err)
	}

	t.Log("output: ", string(result))
}

func TestEntityUpdate(t *testing.T) {
	queryString := "id=test123&source=abcd&type=DEVICE&user_id=admin"
	methodName := "entities"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		fmt.Sprintf("%s?%s", methodName, queryString),
		"PUT",
		&dapr.DataContent{
			ContentType: "application/json",
			Data: []byte(`{
				"zone": "chengdu1",
				"temp": 25
			}`),
		})

	if nil != err {
		t.Log("output: ", err)
	}

	t.Log("output: ", string(result))
}

func TestEntityGET(t *testing.T) {
	queryString := "id=test123&source=abcd&type=DEVICE&user_id=admin"
	methodName := "entities"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		fmt.Sprintf("%s?%s", methodName, queryString),
		"GET",
		&dapr.DataContent{
			ContentType: "application/json",
		})

	if nil != err {
		t.Log("output: ", err)
	}

	t.Log("output: ", string(result))
}

func TestEntityDelete(t *testing.T) {
	queryString := "id=test123&source=abcd&type=DEVICE&user_id=admin"
	methodName := "entities"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		fmt.Sprintf("%s?%s", methodName, queryString),
		"GET",
		&dapr.DataContent{
			ContentType: "application/json",
		})

	if nil != err {
		t.Log("output: ", err)
	}

	t.Log("output: ", string(result))
}

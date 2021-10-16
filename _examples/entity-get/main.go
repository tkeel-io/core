package main

import (
	"context"
	"fmt"

	dapr "github.com/dapr/go-sdk/client"
)

func main() {
	client, err := dapr.NewClient()
	if nil != err {
		panic(err)
	}

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
		panic(err)
	}
	fmt.Println(string(result))
}

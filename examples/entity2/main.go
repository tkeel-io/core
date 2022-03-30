package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tkeel-io/tdtl"
)

type Info struct {
	ID   string
	Path string
}

func main() {
	start := time.Now()
	createNum := 100
	execeptions := []Info{}
	for n := 0; n < createNum; n++ {
		en := createEntity()
		entityID := en.Get("data.deviceObject.id").String()
		res := getEntity(entityID)
		spacePath := res.Get("data.properties.sysField._spacePath").String()
		if !strings.Contains(spacePath, "/") {
			execeptions = append(execeptions, Info{ID: entityID, Path: spacePath})
		}
	}

	elapsed := time.Now().Sub(start).Nanoseconds() / 1e6
	fmt.Println("elapsed: ", elapsed, "elapsed second: ", float64(elapsed)/1000.0/float64(createNum))

	for index := range execeptions {
		fmt.Println("exec:", execeptions[index].ID, execeptions[index].Path)
	}
}

func getEntity(id string) *tdtl.Collect {
	url := fmt.Sprintf("http://192.168.100.8:30342/v1/entities/%s?type=DEVICE&owner=admin&source=CORE", id)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return tdtl.NULL_RESULT
	}
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0ODQwOTI1NCwic3ViIjoidXNyLTNmNDYxNDg5ZDhiZGIxOGI4NWY4NmJlYzEzZjYifQ.Gjom7hEE9P_KojY7xTKt8EExjyfv-dOcDg4zEmQZ2Yfyd3HIjZ1w4QANy4sRnq3K_GpHl9Sj_OTxzEZ32wE9lQ")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return tdtl.NULL_RESULT
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return tdtl.NULL_RESULT
	}

	return tdtl.New(body)
}

func createEntity() *tdtl.Collect {
	url := "http://192.168.100.8:32731/v1/devices"
	method := "POST"

	name := uuid.New().String()
	cc := tdtl.New(`{
		"name": "test1zmxxx",
		"description": "test",
		"parentID": "iotd-99124c67-6611-45b1-93b7-a41ea9dba98b",
		"parentName": "tomas",
		"directConnection": true,
		"templateId": "",
		"selfLearn": false,
		"ext": {
			"location": "wuhan",
			"commany": "qingcloud"
		}
	}`)
	cc.Set("name", tdtl.NewString(name))
	payload := strings.NewReader(cc.String())

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return tdtl.NULL_RESULT
	}
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0ODQwOTI1NCwic3ViIjoidXNyLTNmNDYxNDg5ZDhiZGIxOGI4NWY4NmJlYzEzZjYifQ.Gjom7hEE9P_KojY7xTKt8EExjyfv-dOcDg4zEmQZ2Yfyd3HIjZ1w4QANy4sRnq3K_GpHl9Sj_OTxzEZ32wE9lQ")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return tdtl.NULL_RESULT
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return tdtl.NULL_RESULT
	}

	return tdtl.New(body)
}

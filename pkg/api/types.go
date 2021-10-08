package api

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

const (
	OpenApiSuccessCode = 0

	OpenApiVersion      = "1.0"
	OpenAPISuccessMsg   = "ok"
	OpenApiStatusActive = "ACTIVE"
)

var defaultOpenApiPluginId = "core"

func SetDefaultPluginId(pluginId string) {
	defaultOpenApiPluginId = pluginId
}

type ServicHandler = func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)

type IService interface {
	RegisterService(service common.Service) error
	Name() string
}

//https://cwiki.yunify.com/pages/viewpage.action?pageId=80092844
//IdentifyResponse
type IdentifyResponse struct {
	RetCode  int    `json:"ret"`
	Message  string `json:"msg"`
	PluginId string `json:"plugin_id"`
	Version  string `json:"version"`
}

//StatusResponse
type StatusResponse struct {
	RetCode int    `json:"ret"`
	Message string `json:"msg"`
	Status  string `json:"status"`
}

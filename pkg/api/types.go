package api

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

const (
	OpenAPISuccessCode = 0

	OpenAPIVersion      = "1.0"
	OpenAPISuccessMsg   = "ok"
	OpenAPIStatusActive = "ACTIVE"
)

var defaultOpenAPIPluginID = "core"

func SetDefaultPluginID(pluginID string) {
	defaultOpenAPIPluginID = pluginID
}

type ServiceHandler = func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)

type IService interface {
	RegisterService(service common.Service) error
	Name() string
}

// IdentifyResponse
// this is ref:https://cwiki.yunify.com/pages/viewpage.action?pageId=80092844.
type IdentifyResponse struct {
	Ret      int    `json:"ret"`
	Msg      string `json:"msg"`
	PluginID string `json:"plugin_id"`
	Version  string `json:"version"`
}

type StatusResponse struct {
	Ret    int    `json:"ret"`
	Msg    string `json:"msg"`
	Status string `json:"status"`
}

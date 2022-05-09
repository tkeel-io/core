package service

import (
	go_restful "github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tkeel-io/core/pkg/runtime"
)

type GOPSService struct {
	node *runtime.Node
}

func NewGOPSService() (*GOPSService, error) {
	return &GOPSService{}, nil
}

func (h *GOPSService) Metrics(req *go_restful.Request, resp *go_restful.Response) {
	promhttp.Handler().ServeHTTP(resp, req.Request)
}

func (h *GOPSService) Debug(req *go_restful.Request, resp *go_restful.Response) {
	h.node.Debug(req, resp)
}

func (h *GOPSService) SetNode(instance *runtime.Node) {
	h.node = instance
}

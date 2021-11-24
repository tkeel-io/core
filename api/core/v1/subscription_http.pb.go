// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// protoc-gen-go-http 0.1.0

package v1

import (
	context "context"
	json "encoding/json"
	go_restful "github.com/emicklei/go-restful"
	errors "github.com/tkeel-io/kit/errors"
	http "net/http"
)

import transportHTTP "github.com/tkeel-io/kit/transport/http"

// This is a compile-time assertion to ensure that this generated file
// is compatible with the tkeel package it is being compiled against.
// import package.context.http.go_restful.json.errors.

type SubscriptionHTTPServer interface {
	CreateSubscription(context.Context, *CreateSubscriptionRequest) (*SubscriptionResponse, error)
	DeleteSubscription(context.Context, *DeleteSubscriptionRequest) (*DeleteSubscriptionResponse, error)
	GetSubscription(context.Context, *GetSubscriptionRequest) (*SubscriptionResponse, error)
	ListSubscription(context.Context, *ListSubscriptionRequest) (*ListSubscriptionResponse, error)
	UpdateSubscription(context.Context, *UpdateSubscriptionRequest) (*SubscriptionResponse, error)
}

type SubscriptionHTTPHandler struct {
	srv SubscriptionHTTPServer
}

func newSubscriptionHTTPHandler(s SubscriptionHTTPServer) *SubscriptionHTTPHandler {
	return &SubscriptionHTTPHandler{srv: s}
}

func (h *SubscriptionHTTPHandler) CreateSubscription(req *go_restful.Request, resp *go_restful.Response) {
	in := CreateSubscriptionRequest{}
	if err := transportHTTP.GetBody(req, &in.Subscription); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.CreateSubscription(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteErrorString(httpCode, tErr.Message)
		return
	}

	result, err := json.Marshal(out)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	_, err = resp.Write(result)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *SubscriptionHTTPHandler) DeleteSubscription(req *go_restful.Request, resp *go_restful.Response) {
	in := DeleteSubscriptionRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.DeleteSubscription(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteErrorString(httpCode, tErr.Message)
		return
	}

	result, err := json.Marshal(out)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	_, err = resp.Write(result)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *SubscriptionHTTPHandler) GetSubscription(req *go_restful.Request, resp *go_restful.Response) {
	in := GetSubscriptionRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.GetSubscription(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteErrorString(httpCode, tErr.Message)
		return
	}

	result, err := json.Marshal(out)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	_, err = resp.Write(result)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *SubscriptionHTTPHandler) ListSubscription(req *go_restful.Request, resp *go_restful.Response) {
	in := ListSubscriptionRequest{}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.ListSubscription(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteErrorString(httpCode, tErr.Message)
		return
	}

	result, err := json.Marshal(out)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	_, err = resp.Write(result)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *SubscriptionHTTPHandler) UpdateSubscription(req *go_restful.Request, resp *go_restful.Response) {
	in := UpdateSubscriptionRequest{}
	if err := transportHTTP.GetBody(req, &in.Subscription); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := transportHTTP.GetQuery(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	if err := transportHTTP.GetPathValue(req, &in); err != nil {
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	ctx := transportHTTP.ContextWithHeader(req.Request.Context(), req.Request.Header)

	out, err := h.srv.UpdateSubscription(ctx, &in)
	if err != nil {
		tErr := errors.FromError(err)
		httpCode := errors.GRPCToHTTPStatusCode(tErr.GRPCStatus().Code())
		resp.WriteErrorString(httpCode, tErr.Message)
		return
	}

	result, err := json.Marshal(out)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	_, err = resp.Write(result)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
}

func RegisterSubscriptionHTTPServer(container *go_restful.Container, srv SubscriptionHTTPServer) {
	var ws *go_restful.WebService
	for _, v := range container.RegisteredWebServices() {
		if v.RootPath() == "/v1" {
			ws = v
			break
		}
	}
	if ws == nil {
		ws = new(go_restful.WebService)
		ws.ApiVersion("/v1")
		ws.Path("/v1").Produces(go_restful.MIME_JSON)
		container.Add(ws)
	}

	handler := newSubscriptionHTTPHandler(srv)
	ws.Route(ws.POST("/plugins/{plugin}/subscriptions").
		To(handler.CreateSubscription))
	ws.Route(ws.PUT("/plugins/{plugin}/subscriptions/{id}").
		To(handler.UpdateSubscription))
	ws.Route(ws.DELETE("/plugins/{plugin}/subscriptions/{id}").
		To(handler.DeleteSubscription))
	ws.Route(ws.GET("/plugins/{plugin}/subscriptions/{id}").
		To(handler.GetSubscription))
	ws.Route(ws.GET("/plugins/{plugin}/subscriptions").
		To(handler.ListSubscription))
}

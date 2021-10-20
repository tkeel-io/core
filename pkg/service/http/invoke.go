package http

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/tkeel-io/core/pkg/service"

	"github.com/dapr/go-sdk/service/common"
	"github.com/gorilla/mux"
)

func request2context(r *http.Request, keyList []string) context.Context {
	ctx := context.Background()
	for _, key := range keyList {
		if values := r.Header.Values(key); len(values) > 0 {
			ctx = context.WithValue(ctx, service.ContextKey(key), values[0])
		}
	}
	varsMap := mux.Vars(r)
	plugin := varsMap[service.Plugin]
	entityID := varsMap[service.Entity]
	if entityID != "" {
		ctx = context.WithValue(ctx, service.ContextKey(service.Entity), entityID)
	}
	if plugin != "" {
		ctx = context.WithValue(ctx, service.ContextKey(service.Plugin), plugin)
	}
	return ctx
}

// AddServiceInvocationHandler appends provided service invocation handler with its route to the service.
func (s *Server) AddServiceInvocationHandler(route string, fn func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error {
	var err error
	if route, fn, err = validInvocationEvent(route, fn); err != nil {
		return err
	}

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// capture http args.
			e := &common.InvocationEvent{
				Verb:        r.Method,
				QueryString: r.URL.RawQuery,
				ContentType: r.Header.Get("Content-type"),
			}

			// check for post with no data.
			if r.ContentLength > 0 {
				content, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				e.Data = content
			}

			// execute handler.
			ctx := request2context(r, service.HeaderList)

			o, err := fn(ctx, e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// write to response if handler returned data.
			if o != nil && o.Data != nil {
				if o.ContentType != "" {
					w.Header().Set("Content-type", o.ContentType)
				}
				if _, err := w.Write(o.Data); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		})))

	return nil
}

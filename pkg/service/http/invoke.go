package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tkeel-io/core/pkg/service"

	"github.com/dapr/go-sdk/service/common"
)

func header2context(header http.Header, keyList []string) context.Context {
	ctx := context.Background()
	for _, key := range keyList {
		if values := header.Values(key); len(values) > 0 {
			ctx = context.WithValue(ctx, service.HttpHeader(key), values[0])
		}
	}
	return ctx
}

// AddServiceInvocationHandler appends provided service invocation handler with its route to the service
func (s *Server) AddServiceInvocationHandler(route string, fn func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error {
	if route == "" {
		return fmt.Errorf("service route required")
	}
	if fn == nil {
		return fmt.Errorf("invocation handler required")
	}

	if !strings.HasPrefix(route, "/") {
		route = fmt.Sprintf("/%s", route)
	}

	s.mux.Handle(route, optionsHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// capture http args
			e := &common.InvocationEvent{
				Verb:        r.Method,
				QueryString: r.URL.RawQuery,
				ContentType: r.Header.Get("Content-type"),
			}

			// check for post with no data
			if r.ContentLength > 0 {
				content, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				e.Data = content
			}

			// execute handler
			ctx := header2context(r.Header, service.HeaderList)
			o, err := fn(ctx, e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// write to response if handler returned data
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

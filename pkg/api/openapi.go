package api

import (
	"encoding/json"
	"net/http"
)

func NewOpenApiServeMux() *http.ServeMux {
	serMux := http.NewServeMux()

	handleFunc(serMux, "/v1/identify", identifyHandler)
	handleFunc(serMux, "/v1/status", statusHandler)

	return serMux
}

func identifyHandler(rw http.ResponseWriter, r *http.Request) {

	preDisposeRequest(rw, r)

	in := IdentifyResponse{
		RetCode:  OpenApiSuccessCode,
		Message:  OpenAPISuccessMsg,
		Version:  OpenApiVersion,
		PluginId: defaultOpenApiPluginId,
	}

	bytes, _ := json.Marshal(in)

	_, _ = rw.Write([]byte(bytes))
}

func statusHandler(rw http.ResponseWriter, r *http.Request) {

	preDisposeRequest(rw, r)

	in := StatusResponse{
		RetCode: OpenApiSuccessCode,
		Message: OpenAPISuccessMsg,
		Status:  OpenApiStatusActive,
	}

	bytes, _ := json.Marshal(in)

	_, _ = rw.Write([]byte(bytes))

}

func preDisposeRequest(rw http.ResponseWriter, r *http.Request) {
	for k, values := range r.Header {
		if k == "Authorization" {
			for _, v := range values {
				rw.Header().Add(k, v)
			}
		}
		if k == "x-plugin-jwt" {
			log.Debugf("plugin jwt: %s", values[0])
		}
	}
}

func handleFunc(serMux *http.ServeMux, path string, handler func(rw http.ResponseWriter, r *http.Request)) {
	serMux.HandleFunc(path, handler)
}

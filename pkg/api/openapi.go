package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func NewOpenAPIServeMux() *mux.Router {
	serMux := mux.NewRouter()

	handleFunc(serMux, "/v1/identify", identifyHandler)
	handleFunc(serMux, "/v1/status", statusHandler)

	return serMux
}

func identifyHandler(rw http.ResponseWriter, r *http.Request) {
	preDisposeRequest(rw, r)

	in := IdentifyResponse{
		Ret:      OpenAPISuccessCode,
		Msg:      OpenAPISuccessMsg,
		Version:  OpenAPIVersion,
		PluginID: defaultOpenAPIPluginID,
	}

	bytes, _ := json.Marshal(in)

	if _, err := rw.Write(bytes); err != nil {
		log.Warn(err.Error())
	}
}

func statusHandler(rw http.ResponseWriter, r *http.Request) {
	preDisposeRequest(rw, r)

	in := StatusResponse{
		Ret:    OpenAPISuccessCode,
		Msg:    OpenAPISuccessMsg,
		Status: OpenAPIStatusActive,
	}

	bytes, _ := json.Marshal(in)

	if _, err := rw.Write(bytes); err != nil {
		log.Error(err.Error())
	}
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

func handleFunc(serMux *mux.Router, path string, handler func(rw http.ResponseWriter, r *http.Request)) {
	serMux.HandleFunc(path, handler)
}

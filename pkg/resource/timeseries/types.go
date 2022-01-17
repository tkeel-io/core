package timeseries

import (
	"errors"
	"strings"
)

var (
	ErrInfluxRequiredURL    = errors.New("Influx Error: URL required")
	ErrInfluxRequiredToken  = errors.New("Influx Error: Token required")
	ErrInfluxRequiredOrg    = errors.New("Influx Error: Org required")
	ErrInfluxRequiredBucket = errors.New("Influx Error: Bucket required")
	ErrInfluxInvalidParams  = errors.New("Influx Error: Cannot convert request data")
)

type Engine string

func SwitchToEngine(name string) Engine {
	switch strings.ToLower(name) {
	case "influx", "influxdb", "influxdb2":
		return EngineInflux
	case "noop":
		return EngineNoop
	default:
		return Engine(name)
	}
}

type Data struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]string
	Value       string
	Timestamp   int64
}

type WriteRequest struct {
	Data     interface{}       `json:"data"`
	Metadata map[string]string `json:"metadata"`
}

type QueryRequest interface {
	ToRawQuery() string
}

var _ QueryRequest = RawQueryRequest("")

type RawQueryRequest string

func NewRawQueryRequest(query string) QueryRequest {
	return RawQueryRequest(query)
}

func (r RawQueryRequest) ToRawQuery() string {
	return string(r)
}

type Response struct {
	Raw      []byte            `json:"data"`
	Metadata map[string]string `json:"metadata"`
	Err      error             `json:"error"`
}

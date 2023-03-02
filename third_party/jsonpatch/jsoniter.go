package jsonpatch

import (
	jsonx "encoding/json"
	jsoniter "github.com/json-iterator/go"
)

type JsonRawMessage = jsonx.RawMessage
type JsonSyntaxError = jsonx.SyntaxError

var JsonCompact = jsonx.Compact

var json = jsoniter.ConfigCompatibleWithStandardLibrary

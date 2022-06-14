package util

import (
	jsoniter "github.com/json-iterator/go"
	logf "github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/kit/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func DebugInfo(msg string, v interface{}) {
	log.L().Info(msg)
	bytes, _ := json.Marshal(v)
	log.L().Debug("info: ", logf.String("Values", string(bytes)))
	log.L().Debug("----------------DEBUG----------------")
}

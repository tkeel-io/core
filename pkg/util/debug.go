package util

import (
	"encoding/json"

	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func DebugInfo(msg string, v interface{}) {
	log.Info(msg)
	bytes, _ := json.Marshal(v)
	log.Debug("info: ", zap.String("Values", string(bytes)))
	log.Debug("----------------DEBUG----------------")
}

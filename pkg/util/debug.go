package util

import (
	"encoding/json"

	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

func DebugInfo(msg string, v interface{}) {
	log.L().Info(msg)
	bytes, _ := json.Marshal(v)
	log.L().Debug("info: ", zap.String("Values", string(bytes)))
	log.L().Debug("----------------DEBUG----------------")
}

package util

import (
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

func WrapS(str string) []byte {
	return []byte("\"" + str + "\"")
}

func UnwrapS(bytes []byte) string {
	if len(bytes) > 2 {
		return string(bytes[1 : len(bytes)-1])
	}
	log.Warn("unwrap string failed", zfield.Value(string(bytes)))
	return string(bytes)
}

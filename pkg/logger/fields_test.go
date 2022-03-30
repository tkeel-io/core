package logger

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func Test_Zap(t *testing.T) {
	url := "Hello"
	logger, _ := zap.NewProduction()
	logger.Info("failed to fetch URL",
		zap.String("url", url),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
	logger.Warn("debug log", zap.String("level", url))
	logger.Error("Error Message", zap.String("error", url))
}

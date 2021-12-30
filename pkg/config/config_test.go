package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig(t *testing.T) {
	InitConfig("../../config.yml")
	cfg := GetConfig()
	assert.Equal(t, "core", cfg.Server.AppID)
}

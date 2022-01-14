package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	assert.Equal(t, "", config.Server.AppID)
	InitConfig("../../testdata/testconfig.yml")
	assert.Equal(t, "core", config.Server.AppID)
	assert.Equal(t, "root", config.SearchEngine.ES.Username)
	assert.Equal(t, "root", config.SearchEngine.ES.Password)
	assert.Equal(t, []string{"localhost:8086"}, config.SearchEngine.ES.Urls)
}

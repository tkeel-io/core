package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	assert.Equal(t, "", _config.Server.AppID)
	assert.Equal(t, "", _config.Components.SearchEngine.Use)
	assert.Equal(t, 0, len(_config.Components.SearchEngine.ES.Endpoints))
	assert.Equal(t, "", _config.Components.SearchEngine.ES.Username)
	assert.Equal(t, "", _config.Components.SearchEngine.ES.Password)
	Init("")
	assert.Equal(t, DefaultAppID, _config.Server.AppID)
	assert.Equal(t, _defaultUseSearchEngine, _config.Components.SearchEngine.Use)
	assert.Equal(t, _defaultEtcdConfig.Endpoints, _config.Components.Etcd.Endpoints)
	assert.Equal(t, _defaultESConfig.Username, _config.Components.SearchEngine.ES.Username)
	assert.Equal(t, _defaultESConfig.Password, _config.Components.SearchEngine.ES.Password)
	assert.Equal(t, _defaultESConfig.Endpoints, _config.Components.SearchEngine.ES.Endpoints)
	Init("../../testdata/testconfig.yml")
	assert.Equal(t, "core", _config.Server.AppID)
	assert.Equal(t, "admin", _config.Components.SearchEngine.ES.Username)
	assert.Equal(t, "admin", _config.Components.SearchEngine.ES.Password)
	assert.Equal(t, []string{"http://localhost:9200"}, _config.Components.SearchEngine.ES.Endpoints)
}

func TestAddHTTPScheme(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"localhost", "http://localhost"},
		{"https://localhost", "https://localhost"},
		{"localhost:9200", "http://localhost:9200"},
	}
	for _, test := range tests {
		got := addHTTPScheme(test.s)
		assert.Equal(t, test.want, got)
	}
}

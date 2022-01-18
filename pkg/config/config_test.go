package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	assert.Equal(t, "", _config.Server.AppID)
	assert.Equal(t, "", _config.SearchEngine.Use)
	assert.Equal(t, 0, len(_config.SearchEngine.ES.Address))
	assert.Equal(t, "", _config.SearchEngine.ES.Username)
	assert.Equal(t, "", _config.SearchEngine.ES.Password)
	Init("")
	assert.Equal(t, DefaultAppID, _config.Server.AppID)
	assert.Equal(t, _defaultUseSearchEngine, _config.SearchEngine.Use)
	assert.Equal(t, _defaultEtcdConfig.Address, _config.Etcd.Address)
	assert.Equal(t, _defaultESConfig.Username, _config.SearchEngine.ES.Username)
	assert.Equal(t, _defaultESConfig.Password, _config.SearchEngine.ES.Password)
	assert.Equal(t, _defaultESConfig.Address, _config.SearchEngine.ES.Address)
	Init("../../testdata/testconfig.yml")
	assert.Equal(t, "core", _config.Server.AppID)
	assert.Equal(t, "root", _config.SearchEngine.ES.Username)
	assert.Equal(t, "root", _config.SearchEngine.ES.Password)
	assert.Equal(t, []string{"http://localhost:8086"}, _config.SearchEngine.ES.Address)
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

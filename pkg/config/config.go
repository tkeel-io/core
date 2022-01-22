/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"io/fs"
	"net/url"
	"os"
	"strings"

	"github.com/tkeel-io/core/pkg/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	_httpScheme            = "http"
	_schemeSpliterator     = "://"
	_defaultConfigFilename = "config.yml"
	_corePrefix            = "CORE"

	DefaultAppPort = 6789
	DefaultAppID   = "core"
)

var (
	_config = defaultConfig()
)

var (
	_defaultAppServer = Server{
		AppID:             DefaultAppID,
		AppPort:           DefaultAppPort,
		CoroutinePoolSize: 500,
	}
	_defaultLogConfig = LogConfig{
		Level: "info",
	}
	_defaultUseSearchEngine = "elasticsearch"
	_defaultESConfig        = ESConfig{
		Endpoints: []string{"http://localhost:9200"},
		Username:  "admin",
		Password:  "admin",
	}
	_defaultEtcdConfig = EtcdConfig{
		DialTimeout: 3,
		Endpoints:   []string{"http://localhost:2379"},
	}
)

type Configuration struct {
	Server     Server     `mapstructure:"server"`
	Logger     LogConfig  `mapstructure:"logger"`
	Components Components `mapstructure:"components"`
}

type Components struct {
	Etcd         EtcdConfig   `mapstructure:"etcd"`
	Store        Metadata     `mapstructure:"store"`
	TimeSeries   Metadata     `mapstructure:"time_series"`
	SearchEngine SearchEngine `mapstructure:"search_engine"`
}

type Pair struct {
	Key   string      `yaml:"key"`
	Value interface{} `yaml:"value"`
}

type Metadata struct {
	Name       string `yaml:"name"`
	Properties []Pair `yaml:"properties"`
}

type EtcdConfig struct {
	Endpoints   []string `yaml:"endpoints"`
	DialTimeout int64    `yaml:"dial_timeout"`
}

type SearchEngine struct {
	Use string   `mapstructure:"use" yaml:"use"`
	ES  ESConfig `mapstructure:"elasticsearch" yaml:"elasticsearch"` //nolint:tagliatelle
}

type ESConfig struct {
	Endpoints []string `yaml:"endpoints"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

type LogConfig struct {
	Dev    bool     `yaml:"dev"`
	Level  string   `yaml:"level"`
	Output []string `yaml:"output"`
}

func defaultConfig() Configuration {
	return Configuration{}
}

func Get() Configuration {
	return _config
}

func Init(cfgFile string) {
	if cfgFile != "" {
		// Use Config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search Config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath("./conf")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/core")
	}

	viper.SetEnvPrefix(_corePrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// default.
	viper.SetDefault("server.app_port", DefaultAppPort)
	viper.SetDefault("server.app_id", DefaultAppID)
	viper.SetDefault("server.coroutine_pool_size", _defaultAppServer.CoroutinePoolSize)
	viper.SetDefault("logger.level", _defaultLogConfig.Level)
	viper.SetDefault("components.etcd.endpoints", _defaultEtcdConfig.Endpoints)
	viper.SetDefault("components.etcd.dial_timeout", _defaultEtcdConfig.DialTimeout)
	viper.SetDefault("components.search_engine.use", _defaultUseSearchEngine)
	viper.SetDefault("components.search_engine.elasticsearch.endpoints", _defaultESConfig.Endpoints)
	viper.SetDefault("components.search_engine.elasticsearch.username", _defaultESConfig.Username)
	viper.SetDefault("components.search_engine.elasticsearch.password", _defaultESConfig.Password)

	if err := viper.ReadInConfig(); nil != err {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok || errors.Is(err, fs.ErrNotExist) { //nolint
			// Config file not found.
			defer writeDefault(cfgFile)
		} else {
			panic(err)
		}
	}

	onConfigChanged(fsnotify.Event{Name: "init", Op: fsnotify.Chmod})

	// set callback.
	viper.OnConfigChange(onConfigChanged)
	viper.WatchConfig()
}

func SetEtcdBrokers(brokers []string) {
	for i := 0; i < len(brokers); i++ {
		brokers[i] = addHTTPScheme(brokers[i])
	}
	_config.Components.Etcd.Endpoints = brokers
}

func SetSearchEngineElasticsearchConfig(username, password string, urls []string) {
	for i := 0; i < len(urls); i++ {
		urls[i] = addHTTPScheme(urls[i])
	}

	_config.Components.SearchEngine.ES.Endpoints = urls
	_config.Components.SearchEngine.ES.Username = username
	_config.Components.SearchEngine.ES.Password = password
}

func SetSearchEngineUseDrive(drive string) {
	_config.Components.SearchEngine.Use = drive
}

func onConfigChanged(in fsnotify.Event) {
	_ = viper.Unmarshal(&_config)
	formatEtcdConfigAddr()
	formatESAddress()
}

func formatEtcdConfigAddr() {
	for i := 0; i < len(_config.Components.Etcd.Endpoints); i++ {
		_config.Components.Etcd.Endpoints[i] =
			addHTTPScheme(_config.Components.Etcd.Endpoints[i])
	}
}

func formatESAddress() {
	for i := 0; i < len(_config.Components.SearchEngine.ES.Endpoints); i++ {
		_config.Components.SearchEngine.ES.Endpoints[i] =
			addHTTPScheme(_config.Components.SearchEngine.ES.Endpoints[i])
	}
}

func writeDefault(cfgFile string) {
	if cfgFile == "" {
		cfgFile = _defaultConfigFilename
	}

	if err := viper.WriteConfigAs(cfgFile); nil != err {
		// TODO add write failed handler and remove logger info in this package.
		logger.FailureStatusEvent(os.Stderr, err.Error())
	}
}

func addHTTPScheme(path string) string {
	if strings.Index(path, _schemeSpliterator) > 0 {
		u, err := url.Parse(path)
		if err != nil {
			return path
		}
		if u.Scheme == "" {
			u.Scheme = _httpScheme
		}
		return u.String()
	}
	return _httpScheme + _schemeSpliterator + path
}

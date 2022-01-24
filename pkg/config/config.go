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
	"errors"
	"io/fs"
	"net/url"
	"os"
	"strings"

	"github.com/tkeel-io/core/pkg/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var _config = defaultConfig()

type Configuration struct {
	Server     Server         `yaml:"server" mapstructure:"server"`
	Logger     LogConfig      `yaml:"logger" mapstructure:"logger"`
	Discovery  Discovery      `yaml:"discovery" mapstructure:"discovery"`
	Components Components     `yaml:"components" mapstructure:"components"`
	Dispatcher DispatchConfig `yaml:"dispatcher" mapstructure:"dispatcher"`
}

type Components struct {
	Etcd         EtcdConfig   `yaml:"etcd" mapstructure:"etcd"`
	Store        Metadata     `yaml:"store" mapstructure:"store"`
	Pubsub       Metadata     `yaml:"pubsub" mapstructure:"pubsub"`
	TimeSeries   Metadata     `yaml:"time_series" mapstructure:"time_series"`
	SearchEngine SearchEngine `yaml:"search_engine" mapstructure:"search_engine"`
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
	ES  ESConfig `mapstructure:"elasticsearch" yaml:"elasticsearch"` //nolint
}

type ESConfig struct {
	Endpoints []string `yaml:"endpoints"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

type LogConfig struct {
	Dev      bool     `yaml:"dev"`
	Level    string   `yaml:"level"`
	Output   []string `yaml:"output"`
	Encoding string   `yaml:"encoding"`
}

type Discovery struct {
	Endpoints   []string `yaml:"endpoints"`
	HeartTime   int64    `yaml:"heart_time"`
	DialTimeout int64    `yaml:"dial_timeout"`
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
	viper.SetDefault("server.name", _defaultAppServer.Name)
	viper.SetDefault("server.app_id", _defaultAppServer.AppID)
	viper.SetDefault("server.app_port", _defaultAppServer.AppPort)
	viper.SetDefault("logger.level", _defaultLogConfig.Level)
	viper.SetDefault("logger.output", _defaultLogConfig.Output)
	viper.SetDefault("logger.encoding", _defaultLogConfig.Encoding)
	viper.SetDefault("discovery.endpoints", _defaultDiscovery.Endpoints)
	viper.SetDefault("discovery.heart_time", _defaultDiscovery.HeartTime)
	viper.SetDefault("discovery.dial_timeout", _defaultDiscovery.DialTimeout)
	viper.SetDefault("components.etcd.endpoints", _defaultEtcdConfig.Endpoints)
	viper.SetDefault("components.etcd.dial_timeout", _defaultEtcdConfig.DialTimeout)
	viper.SetDefault("components.search_engine.use", _defaultUseSearchEngine)
	viper.SetDefault("components.search_engine.elasticsearch.endpoints", _defaultESConfig.Endpoints)
	viper.SetDefault("components.search_engine.elasticsearch.username", _defaultESConfig.Username)
	viper.SetDefault("components.search_engine.elasticsearch.password", _defaultESConfig.Password)

	if err := viper.ReadInConfig(); nil != err {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok || errors.As(err, fs.ErrNotExist) { //nolint
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

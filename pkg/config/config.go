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

/*
core 服务配置优先级：
	1. 命令行参数.
	2. 配置文件.
	3. 环境变量.
	4. 默认设置.
*/

package config

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/tkeel-io/core/pkg/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var _cmdViper *viper.Viper
var _config = defaultConfig()

type Configuration struct {
	Proxy      Proxy          `yaml:"proxy" mapstructure:"proxy"`
	Server     Server         `yaml:"server" mapstructure:"server"`
	Logger     LogConfig      `yaml:"logger" mapstructure:"logger"`
	Discovery  Discovery      `yaml:"discovery" mapstructure:"discovery"`
	Components Components     `yaml:"components" mapstructure:"components"`
	Dispatcher DispatchConfig `yaml:"dispatcher" mapstructure:"dispatcher"`
}

type Server struct {
	Name     string   `yaml:"name" mapstructure:"name"`
	AppID    string   `yaml:"app_id" mapstructure:"app_id"`
	HTTPAddr string   `yaml:"http_addr" mapstructure:"http_addr"`
	GRPCAddr string   `yaml:"grpc_addr" mapstructure:"grpc_addr"`
	Sources  []string `yaml:"sources" mapstructure:"sources"`
}

type Proxy struct {
	HTTPPort int `yaml:"http_port" mapstructure:"http_port"`
	GRPCPort int `yaml:"grpc_port" mapstructure:"grpc_port"`
}

type Components struct {
	Etcd         EtcdConfig `yaml:"etcd" mapstructure:"etcd"`
	Store        Metadata   `yaml:"store" mapstructure:"store"`
	TimeSeries   Metadata   `yaml:"time_series" mapstructure:"time_series"`
	SearchEngine string     `yaml:"search_engine" mapstructure:"search_engine"`
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
	Endpoints   []string `yaml:"endpoints" mapstructure:"endpoints"`
	DialTimeout int64    `yaml:"dial_timeout" mapstructure:"dial_timeout"`
}

type LogConfig struct {
	Dev      bool     `yaml:"dev" mapstructure:"dev"`
	Level    string   `yaml:"level" mapstructure:"level"`
	Output   []string `yaml:"output" mapstructure:"output"`
	Encoding string   `yaml:"encoding" mapstructure:"encoding"`
}

type Discovery struct {
	Endpoints   []string `yaml:"endpoints" mapstructure:"endpoints"`
	HeartTime   int64    `yaml:"heart_time" mapstructure:"heart_time"`
	DialTimeout int64    `yaml:"dial_timeout" mapstructure:"dial_timeout"`
}

func defaultConfig() Configuration {
	return Configuration{}
}

func Get() Configuration {
	return _config
}

func GetCmdV() *viper.Viper {
	return _cmdViper
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

	if err := viper.ReadInConfig(); nil != err {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok || errors.As(err, &fs.ErrNotExist) { //nolint
			// Config file not found.
			defer writeDefault(cfgFile)
		} else {
			panic(err)
		}
	}

	onConfigChanged(fsnotify.Event{Name: "init", Op: fsnotify.Chmod})
	// set command line configuration.
	_cmdViper.Unmarshal(&_config)

	viper.WatchConfig()
}

func onConfigChanged(_ fsnotify.Event) {
	_ = viper.Unmarshal(&_config)
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

func init() {
	// default.
	viper.SetDefault("server.name", _defaultAppServer.Name)
	viper.SetDefault("server.app_id", _defaultAppServer.AppID)
	viper.SetDefault("server.http_addr", _defaultAppServer.HTTPAddr)
	viper.SetDefault("server.grpc_addr", _defaultAppServer.GRPCAddr)
	viper.SetDefault("proxy.http_port", _defaultProxyConfig.HTTPPort)
	viper.SetDefault("proxy.grpc_port", _defaultProxyConfig.GRPCPort)
	viper.SetDefault("logger.level", _defaultLogConfig.Level)
	viper.SetDefault("logger.output", _defaultLogConfig.Output)
	viper.SetDefault("logger.encoding", _defaultLogConfig.Encoding)
	viper.SetDefault("discovery.endpoints", _defaultDiscovery.Endpoints)
	viper.SetDefault("discovery.heart_time", _defaultDiscovery.HeartTime)
	viper.SetDefault("discovery.dial_timeout", _defaultDiscovery.DialTimeout)
	viper.SetDefault("components.etcd.endpoints", _defaultEtcdConfig.Endpoints)
	viper.SetDefault("components.etcd.dial_timeout", _defaultEtcdConfig.DialTimeout)

	viper.SetEnvPrefix(_corePrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// initialize command  viper.
	_cmdViper = viper.New()
}

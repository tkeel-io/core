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
	"os"
	"strings"

	"github.com/tkeel-io/core/pkg/print"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tkeel-io/kit/log"
)

const (
	_defaultConfigFilename = "config.yml"
	_corePrefix            = "CORE"

	DefaultAppPort = 6789
	DefaultAppID   = "core"
)

var config = defaultConfig()

type Configuration struct {
	Server     Server     `mapstructure:"server"`
	Logger     LogConfig  `mapstructure:"logger"`
	Etcd       EtcdConfig `mapstructure:"etcd"`
	TimeSeries Metadata   `mapstructure:"time_series"`
}

type Pair struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Metadata struct {
	Name       string `yaml:"name"`
	Properties []Pair `yaml:"properties"`
}

type EtcdConfig struct {
	Address []string
}

type LogConfig struct {
	Dev    bool     `yaml:"dev"`
	Level  string   `yaml:"level"`
	Output []string `yaml:"output"`
}

func defaultConfig() Configuration {
	return Configuration{
		Server: Server{
			AppPort: 6789,
		},
	}
}

func Get() Configuration {
	return config
}

func InitConfig(cfgFile string) {
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

	if err := viper.ReadInConfig(); nil != err {
		if ok := errors.Is(err, viper.ConfigFileNotFoundError{}); ok {
			// Config file not found.
			defer writeDefault(cfgFile)
		} else {
			log.Fatal(err)
		}
	}

	// default.
	viper.SetDefault("server.app_port", DefaultAppPort)
	viper.SetDefault("server.app_id", DefaultAppID)
	viper.SetDefault("server.coroutine_pool_size", 500)
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.output_json", false)
	viper.SetDefault("etcd.address", []string{"http://localhost:2379"})

	// unmarshal
	onConfigChanged(fsnotify.Event{Name: "init", Op: fsnotify.Chmod})

	// set callback.
	viper.OnConfigChange(onConfigChanged)
	viper.WatchConfig()
}

func SetEtcdBrokers(brokers []string) {
	config.Etcd.Address = brokers
}

func onConfigChanged(in fsnotify.Event) {
	_ = viper.Unmarshal(&config)
}

func writeDefault(cfgFile string) {
	if cfgFile == "" {
		cfgFile = _defaultConfigFilename
	}

	if err := viper.WriteConfigAs(cfgFile); nil != err {
		// TODO add write failed handler and remove print info in this package.
		print.FailureStatusEvent(os.Stderr, err.Error())
	}
}

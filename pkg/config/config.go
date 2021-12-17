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
)

var config = defaultConfig()

type Config struct {
	Server Server     `mapstructure:"server"`
	Logger LogConfig  `mapstructure:"logger"`
	Etcd   EtcdConfig `mapstructure:"etcd"`
}

type EtcdConfig struct {
	Address []string
}

type LogConfig struct {
	Dev    bool     `yaml:"dev"`
	Level  string   `yaml:"level"`
	Output []string `yaml:"output"`
}

func defaultConfig() Config {
	return Config{
		Server: Server{
			AppPort: 6789,
		},
	}
}

func GetConfig() *Config {
	return &config
}

func InitConfig(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath("./conf")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/core")
	}

	viper.SetEnvPrefix("CORE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); nil != err {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok { //nolint
			// config file not found.
			defer writeDefault(cfgFile)
		} else {
			panic(errors.Unwrap(err))
		}
	}

	print.InfoStatusEvent(os.Stdout, "loading configuration...")

	// default.
	viper.SetDefault("server.app_port", 6789)
	viper.SetDefault("server.app_id", "core")
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

func onConfigChanged(in fsnotify.Event) {
	_ = viper.Unmarshal(&config)
}

func writeDefault(cfgFile string) {
	if cfgFile == "" {
		cfgFile = "config.yml"
	}

	if err := viper.WriteConfigAs(cfgFile); nil != err {
		// todo...
		print.FailureStatusEvent(os.Stderr, err.Error())
	}
}

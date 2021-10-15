package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/tkeel-io/core/pkg/print"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var config = defaultConfig()

type Config struct {
	Server    Server    `mapstructure:"server"`
	APIConfig APIConfig `mapstructure:"api_config"`
	Logger    LogConfig `mapstructure:"logger"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	OutputJSON bool   `yaml:"output_json"`
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

	// default.
	viper.SetDefault("server.app_port", 6789)
	viper.SetDefault("server.app_id", "core")
	viper.SetDefault("server.coroutine_pool_size", 500)
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.output_json", false)

	// unmarshal
	onConfigChanged(fsnotify.Event{Name: "init", Op: fsnotify.Chmod})

	// set callback.
	viper.OnConfigChange(onConfigChanged)
	viper.WatchConfig()
	print.PendingStatusEvent(os.Stdout, "watch config file.....")
}

func onConfigChanged(in fsnotify.Event) {
	print.PendingStatusEvent(os.Stdout, "watch config event: name(%s), operator(%s).", in.Name, in.Op.String())
	_ = viper.Unmarshal(&config)
	bytes, _ := json.MarshalIndent(config, "	", "	")
	print.InfoStatusEvent(os.Stdout, string(bytes))
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

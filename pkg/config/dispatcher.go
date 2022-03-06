package config

type DispatchConfig struct {
	ID      string   `yaml:"id" mapstructure:"id"`
	Name    string   `yaml:"name" mapstructure:"name"`
	Enabled bool     `yaml:"enabled" mapstructure:"enabled"`
	Sinks   []string `yaml:"sinks" mapstructure:"sinks"`
}

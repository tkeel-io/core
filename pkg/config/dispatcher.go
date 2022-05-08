package config

type DispatchConfig struct {
	ID          string   `yaml:"id" mapstructure:"id"`
	Name        string   `yaml:"name" mapstructure:"name"`
	Enabled     bool     `yaml:"enabled" mapstructure:"enabled"`
	Upstreams   []string `yaml:"upstreams" mapstructure:"upstreams"`
	Downstreams []string `yaml:"downstreams" mapstructure:"downstreams"`
	Logstream   string   `yaml:"logstream" mapstructure:"logstream"`
}

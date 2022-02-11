package config

type DispatchConfig struct {
	ID      string        `yaml:"id" mapstructure:"id"`
	Name    string        `yaml:"name" mapstructure:"name"`
	Enabled bool          `yaml:"enabled" mapstructure:"enabled"`
	Queues  []QueueConfig `yaml:"queues" mapstructure:"queues"`
}

type QueueConfig struct {
	ID           string   `yaml:"id" mapstructure:"id"`
	Name         string   `yaml:"name" mapstructure:"name"`
	Type         string   `yaml:"type" mapstructure:"type"`
	Version      int64    `yaml:"version" mapstructure:"version"`
	NodeName     int64    `yaml:"node_name" mapstructure:"node_name"`
	Consumers    []string `yaml:"consumers" mapstructure:"consumers"`
	ConsumerType string   `yaml:"consumer_type" mapstructure:"consumer_type"`
	Description  string   `yaml:"description" mapstructure:"description"`
	Metadata     []Pair   `yaml:"metadata" mapstructure:"metadata"`
}

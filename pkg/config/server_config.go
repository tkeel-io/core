package config

import (
	"time"
)

type Server struct {
	AppID             string           `mapstructure:"app_id"`
	AppPort           int              `mapstructure:"app_port"`
	CoroutinePoolSize int              `mapstructure:"coroutine_pool_size"`
	TSeriesServers    []*TSeriesServer `mapstructure:"tseries_servers"` //nolint
}

type TSeriesServer struct {
	Name       string     `mapstructure:"name"`
	Enabled    bool       `mapstructure:"enabled"`
	Sources    []Source   `mapstructure:"sources"`
	BatchQueue BatchQueue `mapstructure:"batch_queue"`
}

type BatchQueue struct {
	Name string `mapstructure:"name"`
	// BatchingMaxMessages set the maximum number of messages permitted in a batch. (default: 1000).
	MaxBatching int `mapstructure:"max_batching"`
	// MaxPendingMessages set the max size of the queue.
	MaxPendingMessages uint `mapstructure:"max_pending_messages"`
	// BatchingMaxFlushDelay set the time period within which the messages sent will be batched (default: 10ms).
	BatchingMaxFlushDelay time.Duration `mapstructure:"batching_max_flush_delay"`
}

type Source struct {
	Name       string            `mapstructure:"name"`
	Type       string            `mapstructure:"type"`
	Properties map[string]string `mapstructure:"properties"`
}

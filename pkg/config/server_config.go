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

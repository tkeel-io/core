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
	Name    string   `yaml:"name" mapstructure:"name"`
	AppID   string   `yaml:"app_id" mapstructure:"app_id"`
	AppPort int      `yaml:"app_port" mapstructure:"app_port"`
	Sources []string `yaml:"sources" mapstructure:"sources"`
}

type TSeriesServer struct {
	Name       string     `yaml:"name" mapstructure:"name"`
	Enabled    bool       `yaml:"enabled" mapstructure:"enabled"`
	BatchQueue BatchQueue `yaml:"batch_queue" mapstructure:"batch_queue"`
}

type BatchQueue struct {
	Name string `yaml:"name" mapstructure:"name"`
	// BatchingMaxMessages set the maximum number of messages permitted in a batch. (default: 1000).
	MaxBatching int `yaml:"max_batching" mapstructure:"max_batching"`
	// MaxPendingMessages set the max size of the queue.
	MaxPendingMessages uint `yaml:"max_pending_messages" mapstructure:"max_pending_messages"`
	// BatchingMaxFlushDelay set the time period within which the messages sent will be batched (default: 10ms).
	BatchingMaxFlushDelay time.Duration `yaml:"batching_max_flush_delay" mapstructure:"batching_max_flush_delay"`
}

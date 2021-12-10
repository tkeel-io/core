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

package server

import (
	"context"
	"os"

	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/print"

	"github.com/dapr/go-sdk/service/common"
)

type Manager struct {
	name    string
	service common.Service
	conf    *config.Config

	tseriesServers []*TSeriesServer

	ctx context.Context
}

func NewManager(ctx context.Context, service common.Service, conf *config.Config) *Manager {
	return &Manager{
		ctx:     ctx,
		conf:    conf,
		service: service,
	}
}

func (m *Manager) Init() error {
	// TODO: init event Servers.
	// TODO: init property Servers.
	// TODO: init relationship Servers.

	for _, serverCfg := range m.conf.Server.TSeriesServers {
		if !serverCfg.Enabled {
			print.WarningStatusEvent(os.Stdout, "%s Not Enabled.", serverCfg.Name)
		} else {
			Server := NewTSeriesServer(childContext(m.ctx, m.name), serverCfg.Name, m.service)
			if err := Server.Init(serverCfg); nil != err {
				return err
			}
			m.tseriesServers = append(m.tseriesServers, Server)
		}
	}

	return nil
}

func (m *Manager) Start() error {
	for _, act := range m.tseriesServers {
		if err := act.Run(); nil != err {
			return err
		}
	}
	return nil
}

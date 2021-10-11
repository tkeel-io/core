package server

import (
	"context"
	"fmt"

	"github.com/dapr/go-sdk/service/common"
	"github.com/tkeel-io/core/pkg/config"
)

type ServerManager struct {
	name    string
	service common.Service
	conf    *config.Config

	tseriesServers []*TSeriesServer

	ctx context.Context
}

func NewServerManager(ctx context.Context, service common.Service, conf *config.Config) *ServerManager {
	return &ServerManager{
		ctx:     ctx,
		conf:    conf,
		service: service,
	}
}

func (this *ServerManager) Init() error {

	//init event Servers.

	//init property Servers.

	//init relationship Servers.

	//init tseries srevers.
	for _, serverCfg := range this.conf.Server.TSeriesServers {
		if !serverCfg.Enabled {
			fmt.Printf("%s Not Enabled.", serverCfg.Name)
		} else {
			Server := NewTSeriesServer(childContext(this.ctx, this.name), serverCfg.Name, this.service)
			if err := Server.Init(serverCfg); nil != err {
				return err
			}
			this.tseriesServers = append(this.tseriesServers, Server)
		}
	}

	return nil
}

func (this *ServerManager) Start() error {
	for _, act := range this.tseriesServers {
		if err := act.Run(); nil != err {
			return err
		}
	}
	return nil
}

package proxy

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/runtime/statem"
)

type Proxy struct {
	host         string
	grpcConns    map[string]pb.ProxyClient
	stateManager statem.StateManager
}

func NewProxy(stateManager statem.StateManager) *Proxy {
	return &Proxy{
		host:         "",
		grpcConns:    make(map[string]pb.ProxyClient),
		stateManager: stateManager,
	}
}

func (p *Proxy) RouteMessage(ctx context.Context, msgCtx statem.MessageContext) error {
	var err error
	hostName := ""

	switch hostName {
	case p.host:
		err = p.stateManager.HandleMessage(ctx, msgCtx)
	default:
		proxyClient := p.grpcConns[hostName]
		_, err = proxyClient.Route(ctx, &pb.RouteRequest{
			Header: make(map[string]string),
			Data:   []byte{},
		})
	}

	return errors.Wrap(err, "route message")
}

package proxy

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/core/pkg/util/discovery"
)

type Proxy struct {
	host         string
	grpcConns    map[string]pb.ProxyClient
	stateManager statem.StateManager
	coreResolver *discovery.Resolver
	ctx          context.Context
}

func NewProxy(ctx context.Context, stateManager statem.StateManager, resolver *discovery.Resolver) *Proxy {
	return &Proxy{
		ctx:          ctx,
		host:         "",
		grpcConns:    make(map[string]pb.ProxyClient),
		stateManager: stateManager,
		coreResolver: resolver,
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

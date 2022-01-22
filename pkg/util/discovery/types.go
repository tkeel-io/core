package discovery

import (
	"context"
	"time"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Config struct {
	Endpoints   []string
	DialTimeout int64
}

type Service struct {
	AppID    string            `json:"app_id"`
	Name     string            `json:"name"`
	Host     string            `json:"host"`
	Metadata map[string]string `json:"metadata"`
}

type Register interface {
	Register(ctx context.Context, node Service) error
}

type ResolveHandler func(Service)

type Resolver interface {
	Resolve(ctx context.Context, handlers []ResolveHandler) error
}

type Discovery struct {
	discoveryEnd *clientv3.Client
}

func New(cfg Config) (*Discovery, error) {
	timeout := cfg.DialTimeout * int64(time.Second)
	discoveryEnd, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(timeout),
	})

	if nil != err {
		return nil, errors.Wrap(err, "connect Discovery Endpoint")
	}

	return &Discovery{discoveryEnd: discoveryEnd}, nil
}

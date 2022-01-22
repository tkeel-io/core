package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	GrpcDiscoveryPrefix = "service://"
)

type Config struct {
	Endpoints   []string
	DialTimeout int64
	HeartTime   int64
}

type Service struct {
	AppID    string            `json:"app_id"`
	Name     string            `json:"name"`
	Host     string            `json:"host"`
	Metadata map[string]string `json:"metadata"`
}

func (s Service) Key() string {
	return fmt.Sprintf("%s%s/%s", GrpcDiscoveryPrefix, s.AppID, s.Name)
}

func (s Service) Value() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s Service) WatchKey() string {
	return fmt.Sprintf("%s%s", GrpcDiscoveryPrefix, s.AppID)
}

type Register interface {
	Register(ctx context.Context, node Service) error
}

var (
	PUT    EnventType = EnventType(mvccpb.PUT)
	DELETE EnventType = EnventType(mvccpb.DELETE)
)

type EnventType mvccpb.Event_EventType

type ResolveHandler func(EnventType, Service)

type Resolver interface {
	Resolve(ctx context.Context, handlers []ResolveHandler) error
}

type Discovery struct {
	discoveryEnd *clientv3.Client
	HeartTime    int64
	Config       Config
}

func New(cfg Config) (*Discovery, error) {
	discoveryEnd, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	})

	if nil != err {
		return nil, errors.Wrap(err, "connect Discovery Endpoint")
	}

	return &Discovery{
		Config:       cfg,
		discoveryEnd: discoveryEnd,
		HeartTime:    cfg.HeartTime * int64(time.Second),
	}, nil
}

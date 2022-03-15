package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
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
	AppID    string                 `json:"app_id"`
	Name     string                 `json:"name"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Metadata map[string]interface{} `json:"metadata"`
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
	log.Info("connect on discovery cluster",
		zfield.Endpoints(cfg.Endpoints), zap.Int64("dial_timeout", cfg.DialTimeout))
	discoveryEnd, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	})

	if nil != err {
		log.Error("connect on discovery cluster", zap.Error(err),
			zfield.Endpoints(cfg.Endpoints), zap.Int64("dial_timeout", cfg.DialTimeout))
		return nil, errors.Wrap(err, "connect Discovery Endpoint")
	}

	return &Discovery{
		Config:       cfg,
		discoveryEnd: discoveryEnd,
		HeartTime:    cfg.HeartTime * int64(time.Second),
	}, nil
}

package dapr

import (
	"sync"

	daprSDK "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

var once sync.Once
var pool *daprClientPool

type daprClientPool struct {
	client daprSDK.Client
}

func newPool() *daprClientPool {
	return &daprClientPool{}
}

func (p *daprClientPool) setup() {
	// TODO: !!! daprSDK.NewClient() 可能返回 (nil, nil).
	var err error
	if p.client, err = daprSDK.NewClient(); nil != err {
		log.Error("setup client pool", zap.Error(err))
	}
}

func (p *daprClientPool) Select() daprSDK.Client {
	once.Do(func() {
		p.setup()
	})

	if p.client == nil {
		p.setup()
	}

	return p.client
}

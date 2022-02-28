package dapr

import (
	"sync"

	daprSDK "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/core/pkg/util"
)

var once sync.Once
var pool *daprClientPool

type Client struct {
	id   string
	conn daprSDK.Client
}

func (c *Client) Conn() daprSDK.Client {
	return c.conn
}

type daprClientPool struct {
	size    int
	index   int
	clients []Client
}

func newPool(size int) *daprClientPool {
	return &daprClientPool{
		size:    size,
		clients: make([]Client, size),
	}
}

func (p *daprClientPool) Select() *Client {
	once.Do(func() {
		p.clients = make([]Client, p.size)
		for index := range p.clients {
			cli, err := daprSDK.NewClient()
			if nil == err {
				p.clients[index] =
					Client{id: util.UUID(), conn: cli}
			}
		}
	})

	p.index = (p.index + p.size + 1) % p.size
	return &Client{
		id:   p.clients[p.index].id,
		conn: p.clients[p.index].conn,
	}
}

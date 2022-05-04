package transport

import (
	"context"
	"github.com/tkeel-io/core/pkg/logfield"

	"github.com/tkeel-io/kit/log"
)

var factory = make(map[TransType]Generator)

type Transmitter interface {
	Do(context.Context, *Request) error
	Close() error
}

type Request struct {
	PackageID string
	Method    string
	Address   string
	Header    map[string]string
	Payload   []byte
}

type Generator func() (Transmitter, error)

type TransType string

func (t TransType) String() string {
	return string(t)
}

const (
	TransTypeGRPC TransType = "GRPC"
	TransTypeHTTP TransType = "HTTP"
	TransTypeNOOP TransType = "NOOP"
)

func Register(typ TransType, gen Generator) {
	factory[typ] = gen
}

func New(typ TransType) Transmitter {
	var err error
	var trans Transmitter
	if generator, has := factory[typ]; has {
		if trans, err = generator(); nil == err {
			return trans
		}
		log.L().Error("new Transmitter instance", logf.Error(err), logf.String("type", typ.String()))
	}
	trans, _ = factory["noop"]()
	return trans
}

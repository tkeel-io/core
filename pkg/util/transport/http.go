package transport

import (
	"bytes"
	"context"
	"github.com/tkeel-io/core/pkg/logfield"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	"github.com/tkeel-io/kit/log"
)

var (
	httpIndex  uint32
	clients    []*http.Client
	once       sync.Once
	maxConnect = uint32(6)
)

func init() {
	once.Do(func() {
		clients = make([]*http.Client, maxConnect)
		httpTransport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     500,
		}

		for i := 0; i < int(maxConnect); i++ {
			clients[i] = &http.Client{
				Timeout:   3 * time.Second,
				Transport: httpTransport,
			}
		}
	})
}

type httpTransmitter struct {
	coroutines *ants.Pool
	inited     bool
}

func (tm *httpTransmitter) Do(ctx context.Context, req *Request) error {
	if !tm.inited {
		p, err := ants.NewPool(4000)
		if err != nil {
			log.L().Error("Init httpTransmitter Error", logf.Error(err))
			return xerrors.ErrInvalidHTTPInited
		}
		tm.coroutines = p
		tm.inited = true
	}

	// check request.
	if req.Address == "" {
		log.L().Error("empty target address",
			logf.ID(req.PackageID), logf.Method(req.Method),
			logf.Header(req.Header), logf.Addr(req.Address), logf.Payload(req.Payload))
		return xerrors.ErrInvalidHTTPRequest
	}

	tm.coroutines.Submit(func() { tm.process(req) })

	return nil
}

func (tm *httpTransmitter) process(in *Request) {
	log.L().Debug("delive message through http.Transport",
		logf.ID(in.PackageID), logf.Method(in.Method),
		logf.Header(in.Header), logf.Addr(in.Address), logf.Payload(in.Payload))

	httpReq, _ := http.NewRequest(in.Method, in.Address, bytes.NewBuffer(in.Payload))
	httpReq.Header.Set("Content-Type", "application/json")

	// set header.
	for key, val := range in.Header {
		httpReq.Header.Set(key, val)
	}

	// select client & do request.
	httpCli := clients[httpIndex%maxConnect]
	rsp, err := httpCli.Do(httpReq)
	if nil != err {
		log.L().Error("do http request", logf.Error(err),
			logf.ID(in.PackageID), logf.Method(in.Method),
			logf.Header(in.Header), logf.Addr(in.Address), logf.Payload(in.Payload))
		return
	}

	log.L().Debug("process request completed", logf.ID(in.PackageID),
		logf.Status(rsp.Status), logf.Int("status_code", rsp.StatusCode))

	defer rsp.Body.Close()
	io.Copy(ioutil.Discard, rsp.Body)
}

func (tm *httpTransmitter) Close() error {
	if tm.inited {
		tm.coroutines.Release()
	}
	return nil
}

func init() {
	log.SuccessStatusEvent(os.Stdout, "Register Transmitter<http> successful")
	Register(TransTypeHTTP, func() (Transmitter, error) {
		return &httpTransmitter{}, nil
	})
}

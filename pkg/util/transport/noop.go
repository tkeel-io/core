package transport

import (
	"context"
	"os"

	"github.com/tkeel-io/core/pkg/logfield"
	"github.com/tkeel-io/kit/log"
)

type noopTransmitter struct{}

func (t *noopTransmitter) Do(ctx context.Context, in *Request) error {
	log.L().Debug("delive message through noop.Transport",
		logf.ID(in.PackageID), logf.Method(in.Method),
		logf.Header(in.Header), logf.Addr(in.Address), logf.Payload(in.Payload))
	return nil
}
func (t *noopTransmitter) Close() error {
	return nil
}

func init() {
	log.SuccessStatusEvent(os.Stdout, "Register Transmitter<noop> successful")
	Register(TransTypeNOOP, func() (Transmitter, error) {
		return &noopTransmitter{}, nil
	})
}

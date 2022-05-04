package transport

import (
	"context"
	"os"

	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
)

type noopTransmitter struct{}

func (t *noopTransmitter) Do(ctx context.Context, in *Request) error {
	log.L().Debug("delive message through noop.Transport",
		zfield.ID(in.PackageID), zfield.Method(in.Method),
		zfield.Header(in.Header), zfield.Addr(in.Address), zfield.Payload(in.Payload))
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

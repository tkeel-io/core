package transport

import (
	"context"
	"log"
)

type DaprStateCli interface {
	BatchWrite(ctx context.Context, args *[]interface{}) error
	BuildBulkData(m interface{}) (interface{}, error)
}

func NewDaprStateTransport(ctx context.Context, cli DaprStateCli) (Transport, error) {
	sinkTransport, err := NewSinkTransport(
		ctx,
		"dapr-state",
		func(messages []interface{}) (err error) {
			err = cli.BatchWrite(context.Background(), &messages)
			if err != nil {
				log.Println(err)
			}
			return err
		},
		cli.BuildBulkData)
	return sinkTransport, err
}

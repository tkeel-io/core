/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runtime2

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

type MessageHandle func(ctx context.Context, message interface{}) error

// Sink
type Sink interface {
	String() string
	Send(ctx context.Context, event interface{}) error
	SendAsync(ctx context.Context, event interface{}) (Promise, error)
	Close(ctx context.Context) error
}

// Source
type Source interface {
	String() string
	StartReceiver(ctx context.Context, fn MessageHandle) error
	Close(ctx context.Context) error
}

type Promise interface {
	Then(s func(err error)) Promise
	Finish(err error) Promise
}

type Receiver interface {
	Receive(context.Context, interface{}) error
}

type partitionSource struct {
	urlText           string
	consumer          sarama.Consumer
	partitionConsumer sarama.PartitionConsumer

	ctx    context.Context
	cancel context.CancelFunc
}

func NewPartitionSource(ctx context.Context, source string) Source {
	ctx, cancel := context.WithCancel(ctx)
	return &partitionSource{
		ctx:     ctx,
		cancel:  cancel,
		urlText: source,
	}
}

func (ps *partitionSource) String() string {
	return "partition"
}

func (ps *partitionSource) StartReceiver(ctx context.Context, fn MessageHandle) error {
	urlIns, err := url.Parse(ps.urlText)
	if nil != err {
		return errors.Wrap(err, "parse source")
	}

	brokers := strings.Split(urlIns.Host, ";")
	pathSegs := strings.Split(urlIns.Path, "/")

	var partition int64
	if len(pathSegs) != 3 {
		return errors.New("invalid source")
	} else if partition, err = strconv.ParseInt(pathSegs[2], 10, 64); nil != err {
		return errors.Wrap(err, "parse source partition")
	}

	topic := pathSegs[1]
	consumer, err := sarama.NewConsumer(brokers, nil)
	if nil != err {
		panic(err)
	}

	ps.consumer = consumer
	var pc sarama.PartitionConsumer
	if pc, err = consumer.ConsumePartition(topic,
		int32(partition), sarama.OffsetNewest); err != nil {
		return errors.Wrap(err, "new partition consumer")
	}

	ps.partitionConsumer = pc

	go func(sarama.PartitionConsumer) {
		for {
			select {
			case <-ps.ctx.Done():
				return
			case msg := <-ps.partitionConsumer.Messages():
				fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n",
					msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				fn(context.Background(), msg)
			}
		}
	}(pc)

	return nil
}
func (ps *partitionSource) Close(ctx context.Context) error {
	ps.partitionConsumer.Close()
	ps.consumer.Close()
	ps.cancel()
	return nil
}

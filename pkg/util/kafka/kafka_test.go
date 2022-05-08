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

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/tdtl"
	"sync"
	"testing"
)

type rc struct {
}

func (r *rc) HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var ev = v1.ProtoEvent{}
	if msg == nil || len(msg.Value) == 0 {
		fmt.Println("[-]nil", msg.Topic)
		return nil
	}
	if err := v1.Unmarshal(msg.Value, &ev); nil != err {
		fmt.Println("[-]", msg.Topic, err)
		return err
	}
	if byt, err := json.Marshal(ev); nil != err {
		fmt.Println("[-]", msg.Topic, err)
		return err
	} else {
		fmt.Println("[]", msg.Topic, string(byt))
	}
	return nil
}

func TestNewKafkaPubsub(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c, err := NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core0/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core1/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core2/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core3/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core4/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core5/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core6/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core7/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})
	c, err = NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/core8/dev1")
	t.Log(err)
	c.Received(context.Background(), &rc{})

	wg.Wait()
}

type rc2 struct {
}

func (r *rc2) HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	cc := tdtl.New(msg.Value)
	fmt.Println("[]", msg.Topic, cc.Get(""))

	return nil
}
func TestNewKafkaPubsub2(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c, err := NewKafkaPubsub("kafka://tkeel-middleware-kafka:9092/log0/core")
	t.Log(err)
	c.Received(context.Background(), &rc2{})

	wg.Wait()
}

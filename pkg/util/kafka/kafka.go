package kafka

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	logf "github.com/tkeel-io/core/pkg/logfield"

	"github.com/Shopify/sarama"
	"github.com/dapr/kit/retry"
	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/kit/log"
)

type kafkaMetadata struct {
	Topic   string   `json:"topic" mapstructure:"topic"`
	Group   string   `json:"group" mapstructure:"group"`
	Brokers []string `json:"brokers" mapstructure:"brokers"`
	Timeout int64    `json:"timeout" mapstructure:"timeout"`
}

func parseURL(sink string) (*kafkaMetadata, error) {
	urlIns, err := url.Parse(sink)
	if nil != err {
		return nil, errors.Wrap(err, "parse sink")
	}

	segs := strings.Split(urlIns.Path, "/")
	if len(segs) != 3 {
		return nil, errors.New("invalid sink")
	}

	return &kafkaMetadata{
		Topic:   segs[1],
		Group:   segs[2],
		Brokers: strings.Split(urlIns.Host, ","),
	}, nil
}

func NewKafkaPubsub(urlText string) (*Pubsub, error) {
	var (
		err       error
		client    sarama.Client
		kafkaMeta *kafkaMetadata
		producer  sarama.SyncProducer
	)

	if kafkaMeta, err = parseURL(urlText); nil != err {
		return nil, errors.Wrap(err, "decode pubsub.kafka configuration")
	}

	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Version = sarama.V2_3_0_0
	kafkaCfg.Producer.Retry.Max = 3
	kafkaCfg.Producer.RequiredAcks = sarama.WaitForAll
	kafkaCfg.Producer.Return.Successes = true
	if client, err = sarama.NewClient(kafkaMeta.Brokers, kafkaCfg); nil != err {
		return nil, errors.Wrap(err, "create kafka client instance")
	} else if producer, err = sarama.NewSyncProducerFromClient(client); nil != err {
		return nil, errors.Wrap(err, "create kafka producer instance")
	}

	return &Pubsub{
		id:            kafkaMeta.Topic,
		kafkaClient:   client,
		kafkaMetadata: kafkaMeta,
		kafkaProducer: producer,
	}, nil
}

type Pubsub struct {
	id            string
	kafkaClient   sarama.Client
	kafkaConsumer sarama.ConsumerGroup
	kafkaProducer sarama.SyncProducer
	kafkaMetadata *kafkaMetadata
}

func (k *Pubsub) ID() string {
	return k.id
}

func (k *Pubsub) Send(ctx context.Context, event v1.Event) error {
	if event == nil {
		return nil
	}
	log.L().Debug("pubsub.kafka send", logf.Message(event), logf.Topic(k.kafkaMetadata.Topic),
		logf.ID(k.id), logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))

	if bytes, err := v1.Marshal(event); nil != err {
		log.L().Error("encode payload", logf.Error(err), logf.ID(k.id),
			logf.Topic(k.kafkaMetadata.Topic), logf.Any("event", event))
	} else {
		k.SendBytes(ctx, bytes)
	}
	return nil
}
func (k *Pubsub) SendBytes(ctx context.Context, bytes []byte) error {
	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(""),
		Topic: k.kafkaMetadata.Topic,
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err := k.kafkaProducer.SendMessage(msg)
	return errors.Wrap(err, "kafka client send message")
}

type KafkaReceiver interface { //nolint
	HandleMessage(context.Context, *sarama.ConsumerMessage) error
}

func (k *Pubsub) Received(ctx context.Context, receiver KafkaReceiver) error {
	c, err := sarama.NewConsumerGroupFromClient(k.kafkaMetadata.Group, k.kafkaClient)
	if nil != err {
		log.L().Error("create group consumer instance", logf.ID(k.id), logf.Topic(k.kafkaMetadata.Topic),
			logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))
		return errors.Wrap(err, "create group consumer instance")
	}

	k.kafkaConsumer = c
	log.L().Debug("start receive", logf.ID(k.id), logf.Topic(k.kafkaMetadata.Topic),
		logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))

	go func() {
		defer func() {
			log.L().Debug("Closing ConsumerGroup for topics", logf.Topic(k.kafkaMetadata.Topic),
				logf.ID(k.id), logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))
			if err := k.kafkaConsumer.Close(); err != nil {
				log.L().Error("Error closing consumer group", logf.Error(err), logf.Topic(k.kafkaMetadata.Topic),
					logf.ID(k.id), logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))
			}
		}()

		log.L().Debug("Subscribed and listening to topics", logf.Topic(k.kafkaMetadata.Topic),
			logf.ID(k.id), logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))

		for {
			// Consume the requested topic.
			if innerError := k.kafkaConsumer.Consume(ctx, []string{k.kafkaMetadata.Topic}, &kafkaConsumer{receiver: receiver}); innerError != nil {
				log.L().Error("Error closing consumer group", logf.Error(innerError), logf.Topic(k.kafkaMetadata.Topic),
					logf.ID(k.id), logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))
			}

			if ctx.Err() != nil {
				log.L().Error("Context error, stopping consumer", logf.Error(ctx.Err()), logf.Topic(k.kafkaMetadata.Topic),
					logf.ID(k.id), logf.Endpoints(k.kafkaMetadata.Brokers), logf.Group(k.kafkaMetadata.Group))
				return
			}
		}
	}()
	return nil
}

func (k *Pubsub) Commit(v interface{}) error {
	return nil
}

func (k *Pubsub) Close() error {
	log.L().Info("pubsub.noop close", logf.ID(k.id))
	return nil
}

type kafkaConsumer struct {
	receiver KafkaReceiver
}

func (consumer *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	if consumer.receiver == nil {
		return fmt.Errorf("nil consumer callback")
	}

	backOffConfig := retry.Config{}
	b := backOffConfig.NewBackOffWithContext(session.Context())
	for msg := range claim.Messages() {
		if err := retry.NotifyRecover(func() error {
			var innerErr error
			log.L().Debug("processing kafka message", logf.Topic(msg.Topic),
				logf.Partition(msg.Partition), logf.Offset(msg.Offset), logf.Key(string(msg.Key)))
			if innerErr = consumer.receiver.HandleMessage(session.Context(), msg); innerErr == nil {
				session.MarkMessage(msg, "")
			}
			log.L().Debug("processing kafka message", logf.Topic(msg.Topic),
				logf.Partition(msg.Partition), logf.Offset(msg.Offset), logf.Key(string(msg.Key)))
			return errors.Wrap(innerErr, "handle message")
		}, b, func(err error, d time.Duration) {
			log.L().Debug("processing kafka message", logf.Topic(msg.Topic),
				logf.Partition(msg.Partition), logf.Offset(msg.Offset), logf.Key(string(msg.Key)))
		}, func() {
			log.L().Debug("processing kafka message", logf.Topic(msg.Topic),
				logf.Partition(msg.Partition), logf.Offset(msg.Offset), logf.Key(string(msg.Key)))
		}); err != nil {
			log.L().Error("processing kafka message", logf.Topic(msg.Topic),
				logf.Partition(msg.Partition), logf.Offset(msg.Offset), logf.Key(string(msg.Key)))
			return errors.Wrap(err, "handle message")
		}
	}

	return nil
}

func (consumer *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *kafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

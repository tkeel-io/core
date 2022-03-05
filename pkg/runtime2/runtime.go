package runtime3

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type SourceConf struct {
	Topics    []string
	Brokers   []string
	GroupName string
}

type RuntimeConfig struct {
	Source SourceConf
}

type Runtime struct {
	containers map[string]*Container
	dispatch   Dispatcher
	repo       repository.IRepository

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRuntime(ctx context.Context, repo repository.IRepository, dispatcher Dispatcher) *Runtime {
	ctx, cacel := context.WithCancel(ctx)
	return &Runtime{
		ctx:        ctx,
		cancel:     cacel,
		repo:       repo,
		dispatch:   dispatcher,
		containers: make(map[string]*Container),
	}
}

func (r *Runtime) Start(cfg RuntimeConfig) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	gconsumer, err := sarama.NewConsumerGroup(cfg.Source.Brokers, cfg.Source.GroupName, config)
	if err != nil {
		log.Error("create consumer", zfield.Endpoints(cfg.Source.Brokers), zap.Error(err))
		return errors.Wrap(err, "create consumer")
	}

	consumerProxy := newConsumers(r.ctx, r)
	if err = gconsumer.Consume(r.ctx, cfg.Source.Topics, consumerProxy); nil != err {
		log.Error("consume source", zfield.Endpoints(cfg.Source.Brokers), zap.Error(err))
		return errors.Wrap(err, "consume source")
	}

	return nil
}

func (r *Runtime) GetContainer(id string) *Container {
	if _, has := r.containers[id]; !has {
		log.Info("create container", zfield.ID(id))
		r.containers[id] = NewContainer(r.ctx, id)
	}

	return r.containers[id]
}

type runtimeConsumer struct {
	runtime *Runtime
	ctx     context.Context
}

func newConsumers(ctx context.Context, r *Runtime) *runtimeConsumer {
	return &runtimeConsumer{ctx: ctx, runtime: r}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (rc *runtimeConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (rc *runtimeConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (rc *runtimeConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for {
		select {
		case <-rc.ctx.Done():
			return nil
		case msg := <-claim.Messages():
			log.Debug("consume message", zfield.Offset(msg.Offset),
				zfield.Partition(msg.Partition), zfield.Topic(msg.Topic), zap.Any("header", msg.Headers))

			containerID := fmt.Sprintf("%s-%d", msg.Topic, msg.Partition)
			container := rc.runtime.GetContainer(containerID)
			container.DeliveredEvent(context.Background(), msg)
		}
	}
}

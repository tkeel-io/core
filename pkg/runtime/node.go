package runtime

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

type NodeConf struct {
	Source SourceConf
}

type Node struct {
	runtimes map[string]*Runtime
	dispatch Dispatcher
	repo     repository.IRepository

	ctx    context.Context
	cancel context.CancelFunc
}

func NewNode(ctx context.Context, repo repository.IRepository, dispatcher Dispatcher) *Node {
	ctx, cacel := context.WithCancel(ctx)
	return &Node{
		ctx:      ctx,
		cancel:   cacel,
		repo:     repo,
		dispatch: dispatcher,
		runtimes: make(map[string]*Runtime),
	}
}

func (n *Node) Start(cfg NodeConf) error {
	log.Info("start node...")

	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	gconsumer, err := sarama.NewConsumerGroup(cfg.Source.Brokers, cfg.Source.GroupName, config)
	if err != nil {
		log.Error("create consumer", zfield.Endpoints(cfg.Source.Brokers), zap.Error(err))
		return errors.Wrap(err, "create consumer")
	}

	consumerProxy := newConsumers(n.ctx, n)
	if err = gconsumer.Consume(n.ctx, cfg.Source.Topics, consumerProxy); nil != err {
		log.Error("consume source", zfield.Endpoints(cfg.Source.Brokers), zap.Error(err))
		return errors.Wrap(err, "consume source")
	}

	return nil
}

func (n *Node) HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) {
	rid := fmt.Sprintf("%s-%d", msg.Topic, msg.Partition)
	if _, has := n.runtimes[rid]; !has {
		log.Info("create container", zfield.ID(rid))
		n.runtimes[rid] = NewRuntime(n.ctx, rid)
	}

	// load runtime spec.
	runtime := n.runtimes[rid]
	runtime.DeliveredEvent(context.Background(), msg)
}

type nodeConsumer struct {
	node *Node
	ctx  context.Context
}

func newConsumers(ctx context.Context, node *Node) *nodeConsumer {
	return &nodeConsumer{ctx: ctx, node: node}
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *nodeConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (c *nodeConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (c *nodeConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		case msg := <-claim.Messages():
			log.Debug("consume message", zfield.Offset(msg.Offset),
				zfield.Partition(msg.Partition), zfield.Topic(msg.Topic), zap.Any("header", msg.Headers))
			c.node.HandleMessage(context.Background(), msg)
		}
	}
}

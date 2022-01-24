package dispatch

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/statem"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type dispatcher struct {
	ID                    string
	Name                  string
	upstreamQueues        map[string]*dao.Queue
	downstreamQueues      map[string]*dao.Queue
	upstreamConnections   map[string]pubsub.Pubsub
	downstreamConnections map[string]pubsub.Pubsub
	coreRepository        repository.IRepository
	hashTable             sort.StringSlice

	ctx    context.Context
	cancel context.CancelFunc
}

func NewDispatcher(ctx context.Context, id string, name string, repo repository.IRepository) Dispatcher {
	ctx, cancel := context.WithCancel(ctx)
	return &dispatcher{
		ctx:                   ctx,
		cancel:                cancel,
		ID:                    id,
		Name:                  name,
		coreRepository:        repo,
		upstreamQueues:        make(map[string]*dao.Queue),
		downstreamQueues:      make(map[string]*dao.Queue),
		upstreamConnections:   make(map[string]pubsub.Pubsub),
		downstreamConnections: make(map[string]pubsub.Pubsub),
	}
}

func (d *dispatcher) Run() error {
	log.Info("run dispatcher",
		zfield.ID(d.ID), zfield.Name(d.Name))

	var err error
	// setup queues.
	if err = d.setup(); nil != err {
		log.Error("setup dispatcher", zap.Error(err),
			zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
		return errors.Wrap(err, "setup dispatcher")
	}

	// consume queues.
	for id, pubsubInstance := range d.upstreamConnections {
		log.Info("receive pubsub", zfield.ID(id),
			zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
		if err = pubsubInstance.Received(context.Background(),
			func(ctx context.Context, message interface{}) error {
				msgCtx, _ := message.(statem.MessageContext)
				entityID := msgCtx.Headers.GetReceiver()

				selector := util.Hash32(entityID)
				selectIndex := selector % uint32(d.hashTable.Len())

				log.Debug("dispatch pubsub message", zfield.Eid(entityID), zap.String("select_queue", d.hashTable[selectIndex]))
				return nil
			}); nil != err {
			log.Error("start receive pubsub", zfield.ID(id),
				zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
		}
	}

	return nil
}

func (d *dispatcher) Stop() error {
	log.Info("stop dispatcher", zfield.ID(d.ID), zfield.Name(d.Name))

	for id, pubsubConn := range d.upstreamConnections {
		pubsubConn.Close()
		log.Debug("stop upstream queue", zfield.ID(id),
			zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
	}
	for id, pubsubConn := range d.downstreamConnections {
		pubsubConn.Close()
		log.Debug("stop downstream queue", zfield.ID(id),
			zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
	}
	return nil
}

func (d *dispatcher) constructQueue(queue *dao.Queue) {
	pubsubInst := pubsub.NewPubsub(resource.Metadata{
		Name:       queue.ConsumerType.String(),
		Properties: queue.Metadata,
	})

	d.upstreamConnections[queue.ID] = pubsubInst

	var fmtString string
	switch queue.ConsumerType {
	case dao.ConsumerTypeCore:
		fmtString = "initialize downstream queue"
		d.appendDownstream(queue.ID)
		d.downstreamQueues[queue.ID] = queue
		d.downstreamConnections[queue.ID] = pubsubInst
	case dao.ConsumerTypeDispatch:
		fmtString = "initialize upstream queue"
		d.removeDownstream(queue.ID)
		d.upstreamQueues[queue.ID] = queue
		d.upstreamConnections[queue.ID] = pubsubInst
	default:
		log.Error("Queue consumer type unknown", zfield.DispatcherID(d.ID),
			zfield.DispatcherName(d.Name), zap.Error(xerrors.ErrInvalidQueueConsumerType))
	}

	log.Info(fmtString, zfield.ID(queue.ID),
		zfield.Name(queue.Name), zfield.Desc(queue.Description),
		zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name),
		zap.String("consumer_type", queue.ConsumerType.String()),
		zfield.Type(queue.Type.String()), zfield.Version(queue.Version),
		zap.Any("metadata", queue.Metadata), zap.Strings("consumers", queue.Consumers))
}

func (d *dispatcher) closeQueue(queue *dao.Queue) {
	var (
		fmtString        string
		closeQueues      map[string]*dao.Queue
		closeConnections map[string]pubsub.Pubsub
	)
	switch queue.ConsumerType {
	case dao.ConsumerTypeCore:
		fmtString = "close downstream queue"
		closeQueues = d.downstreamQueues
		closeConnections = d.downstreamConnections
	case dao.ConsumerTypeDispatch:
		fmtString = "close upstream queue"
		closeQueues = d.upstreamQueues
		closeConnections = d.upstreamConnections
	default:
		// never.
	}

	// close queue if exists.
	if _, exist := closeConnections[queue.ID]; exist {
		if err := closeConnections[queue.ID].Close(); nil != err {
			log.Error("close queue", zfield.ID(queue.ID),
				zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
		}

		// clean queue.
		delete(closeQueues, queue.ID)
		delete(closeConnections, queue.ID)
	}

	log.Info(fmtString, zfield.ID(queue.ID),
		zfield.Name(queue.Name), zfield.Desc(queue.Description),
		zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name),
		zap.String("consumer_type", queue.ConsumerType.String()),
		zfield.Type(queue.Type.String()), zfield.Version(queue.Version),
		zap.Any("metadata", queue.Metadata), zap.Strings("consumers", queue.Consumers))
}

func (d *dispatcher) setup() error {
	// list current queues.
	elapsedTime := util.NewElapsed()
	d.coreRepository.RangeQueue(context.Background(), 0, func(queues []dao.Queue) {
		for index := range queues {
			// construct queue.
			d.constructQueue(&queues[index])
		}
	})

	log.Info("initialize queues", zfield.ID(d.ID),
		zfield.Name(d.Name), zfield.Elapsed(elapsedTime.Elapsed()))

	// watch queue modify.
	go d.coreRepository.WatchQueue(context.Background(), 0, func(et dao.EnventType, queue dao.Queue) {
		switch et {
		case dao.PUT:
			d.closeQueue(&queue)
			d.constructQueue(&queue)
		case dao.DELETE:
			d.closeQueue(&queue)
		default:
			log.Error("invalid EventType", zfield.Name(queue.Name),
				zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name),
				zfield.Type(queue.Type.String()), zfield.Version(queue.Version),
				zfield.Desc(queue.Description), zap.String("event_type", et.String()),
				zap.String("consumer_type", queue.ConsumerType.String()), zfield.ID(queue.ID),
				zap.Any("metadata", queue.Metadata), zap.Strings("consumers", queue.Consumers))
		}
	})

	return nil
}

func (d *dispatcher) appendDownstream(queueID string) {
	d.hashTable = append(d.hashTable, queueID)
	sort.Sort(d.hashTable)
}

func (d *dispatcher) removeDownstream(queueID string) {
	index := d.hashTable.Search(queueID)
	if index < d.hashTable.Len() && queueID == d.hashTable[index] {
		d.hashTable = append(d.hashTable[:index], d.hashTable[index+1:]...)
	}
}

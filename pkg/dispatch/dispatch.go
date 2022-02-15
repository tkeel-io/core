package dispatch

import (
	"context"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/core/pkg/runtime/message"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/core/pkg/util/transport"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type dispatcher struct {
	ID                    string
	Name                  string
	Enabled               bool
	upstreamQueues        map[string]*dao.Queue
	downstreamQueues      map[string]*dao.Queue
	upstreamConnections   map[string]pubsub.Pubsub
	downstreamConnections map[string]pubsub.Pubsub
	coreRepository        repository.IRepository
	loopbackConnection    pubsub.Pubsub
	transmitter           transport.Transmitter

	ctx    context.Context
	cancel context.CancelFunc
}

func New(ctx context.Context, id string, name string, enabled bool, repo repository.IRepository) *dispatcher { //nolint
	ctx, cancel := context.WithCancel(ctx)
	transType := transport.TransTypeHTTP

	return &dispatcher{
		ID:                    id,
		ctx:                   ctx,
		cancel:                cancel,
		Name:                  name,
		Enabled:               enabled,
		coreRepository:        repo,
		transmitter:           transport.New(transType),
		upstreamQueues:        make(map[string]*dao.Queue),
		downstreamQueues:      make(map[string]*dao.Queue),
		upstreamConnections:   make(map[string]pubsub.Pubsub),
		downstreamConnections: make(map[string]pubsub.Pubsub),
	}
}

func (d *dispatcher) Dispatch(ctx context.Context, ev cloudevents.Event) error {
	var err error
	var data []byte
	var msgType string
	var callbackEnd string
	ev.ExtensionAs(message.ExtMessageType, &msgType)
	ev.ExtensionAs(message.ExtCallback, &callbackEnd)

	if data, err = ev.DataBytes(); nil != err {
		log.Error("get data", zap.Error(err),
			zfield.ID(ev.ID()), zfield.Header(message.GetAttributes(ev)))
		return errors.Wrap(err, "dispatch event")
	}

	switch message.MessageType(msgType) {
	case message.MessageTypeRespond:
		log.Debug("dispatch callback", zfield.ID(ev.ID()),
			zfield.Header(message.GetAttributes(ev)))
		d.transmitter.Do(ctx, &transport.Request{
			PackageID: ev.ID(),
			Method:    http.MethodPost,
			Address:   callbackEnd,
			Header:    message.GetAttributes(ev),
			Payload:   data,
		})
	default:
		if err = d.loopbackConnection.Send(ctx, ev); nil != err {
			log.Error("dispatch event", zap.Error(err), zfield.Event(ev),
				zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
		}
	}

	return nil
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
		d.startConsumeQueue(id, pubsubInstance)
	}

	return nil
}

func (d *dispatcher) startConsumeQueue(id string, pubsubIns pubsub.Pubsub) {
	log.Info("pubsub start receive", zfield.ID(id),
		zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))

	var err error
	if err = pubsubIns.Received(context.Background(),
		func(ctx context.Context, ev cloudevents.Event) error {
			var entityID string
			ev.ExtensionAs(message.ExtEntityID, &entityID)
			selectQueue := placement.Global().Select(entityID)
			selectConn := d.downstreamConnections[selectQueue.ID]
			// append som attributes.
			ev.SetExtension(message.ExtChannelID, id)

			// send event.
			if err = selectConn.Send(ctx, ev); nil != err {
				log.Error("dispatch message", zfield.Eid(entityID),
					zap.String("select_queue", selectQueue.ID))
			}

			log.Debug("dispatch pubsub message",
				zfield.Eid(entityID), zap.Any("select_queue", selectQueue))

			return nil
		}); nil != err {
		log.Error("start receive pubsub", zfield.ID(id),
			zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name))
	}
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

	// reset.
	d.upstreamQueues = make(map[string]*dao.Queue)
	d.upstreamConnections = make(map[string]pubsub.Pubsub)
	d.downstreamQueues = make(map[string]*dao.Queue)
	d.downstreamConnections = make(map[string]pubsub.Pubsub)
	return nil
}

func (d *dispatcher) constructQueues(queues []dao.Queue) {
	for index := range queues {
		switch queues[index].ConsumerType {
		case dao.ConsumerTypeCore:
			d.constructDownstreamQueue(&queues[index])
		case dao.ConsumerTypeDispatch:
			if d.Enabled {
				d.constructUpstreamQueue(&queues[index])
			}
		default:
			log.Error("invalid consumer type",
				zfield.ID(queues[index].ID), zfield.Type(queues[index].ConsumerType.String()))
		}
	}
}

func (d *dispatcher) constructUpstreamQueue(queue *dao.Queue) pubsub.Pubsub {
	pubsubInst := pubsub.NewPubsub(queue.ID, resource.Metadata{
		Name:       queue.Type.String(),
		Properties: queue.Metadata,
	})

	d.upstreamQueues[queue.ID] = queue
	d.upstreamConnections[queue.ID] = pubsubInst

	log.Info("initialize upstream queue", zfield.ID(queue.ID),
		zfield.Name(queue.Name), zfield.Desc(queue.Description),
		zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name),
		zap.String("consumer_type", queue.ConsumerType.String()),
		zfield.Type(queue.Type.String()), zfield.Version(queue.Version),
		zap.Any("metadata", queue.Metadata), zap.Strings("consumers", queue.Consumers))

	return pubsubInst
}

func (d *dispatcher) constructDownstreamQueue(queue *dao.Queue) pubsub.Pubsub {
	pubsubInst := pubsub.NewPubsub(queue.ID, resource.Metadata{
		Name:       queue.Type.String(),
		Properties: queue.Metadata,
	})

	d.downstreamQueues[queue.ID] = queue
	placement.Global().AppendQueue(*queue)
	d.downstreamConnections[queue.ID] = pubsubInst

	log.Info("initialize downstream queue", zfield.ID(queue.ID),
		zfield.Name(queue.Name), zfield.Desc(queue.Description),
		zfield.DispatcherID(d.ID), zfield.DispatcherName(d.Name),
		zap.String("consumer_type", queue.ConsumerType.String()),
		zfield.Type(queue.Type.String()), zfield.Version(queue.Version),
		zap.Any("metadata", queue.Metadata), zap.Strings("consumers", queue.Consumers))

	return pubsubInst
}

func (d *dispatcher) closeQueue(queue *dao.Queue) {
	var (
		fmtString        string
		closeQueues      map[string]*dao.Queue
		closeConnections map[string]pubsub.Pubsub
	)
	switch queue.ConsumerType {
	case dao.ConsumerTypeCore:
		log.Warn("access ConsumerTypeDispatch queue updated", zfield.ID(queue.ID))
		return
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
	// setpu loopback Queue.
	d.constructUpstreamQueue(loopbackQueue)
	d.loopbackConnection = d.upstreamConnections[loopbackQueue.ID]

	// list current queues.
	elapsedTime := util.NewElapsed()
	revision := d.coreRepository.GetLastRevision(context.Background())
	d.coreRepository.RangeQueue(context.Background(), revision,
		func(queues []dao.Queue) {
			d.constructQueues(queues)
		})

	log.Info("initialize queues", zfield.ID(d.ID),
		zfield.Name(d.Name), zfield.Elapsed(elapsedTime.Elapsed()))

	// watch queue modify.
	revision = d.coreRepository.GetLastRevision(context.Background())
	go d.coreRepository.WatchQueue(context.Background(), revision, func(et dao.EnventType, queue dao.Queue) {
		log.Info("catch an event", zfield.Type(et.String()),
			zap.String("queue_id", queue.ID), zap.String("queue_name", queue.Name))
		switch et {
		case dao.PUT:
			switch queue.ConsumerType {
			case dao.ConsumerTypeDispatch:
				d.closeQueue(&queue)
				pubsubIns := d.constructUpstreamQueue(&queue)
				d.startConsumeQueue(pubsubIns.ID(), pubsubIns)
			default:
				log.Warn("access ConsumerTypeDispatch queue updated", zfield.ID(queue.ID))
			}
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

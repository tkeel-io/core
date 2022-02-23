package runtime

import (
	"context"
	"time"

	"github.com/tkeel-io/core/pkg/config"
	"github.com/tkeel-io/core/pkg/inbox"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/resource"
	"github.com/tkeel-io/core/pkg/resource/pubsub"
	"github.com/tkeel-io/kit/log"
)

func (m *Manager) initializeSources() {
	m.listQueue()
	go m.watchQueue()
}

func (m *Manager) listQueue() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()

	repo := m.Resource().Repo()
	revision := repo.GetLastRevision(ctx)
	coreNodeName := config.Get().Server.Name

	repo.RangeQueue(context.Background(), revision, func(queues []dao.Queue) {
		// create receiver.
		for _, queue := range queues {
			switch queue.ConsumerType {
			case dao.ConsumerTypeCore:
				if coreNodeName == queue.NodeName {
					log.Info("append queue", zfield.ID(queue.ID))
					// create receiver instance.
					receiver := pubsub.NewPubsub(queue.ID, resource.Metadata{
						Name:       queue.Type.String(),
						Properties: queue.Metadata,
					})

					inboxIns := inbox.New(m.ctx, receiver)
					if _, has := m.inboxes[queue.ID]; has {
						m.inboxes[queue.ID].Close()
					}
					m.inboxes[queue.ID] = inboxIns
				}
			default:
			}
		}
	})

	// start consumer inbox.
	for id, inboxIns := range m.inboxes {
		log.Info("start consumer inbox", zfield.ID(id))
		inboxIns.Consume(m.ctx, m.HandleMessage)
		m.containers[id] = NewContainer(m.ctx, id, m)
	}
}

func (m *Manager) watchQueue() {
	ctx, cancel := context.WithTimeout(m.ctx, 3*time.Second)
	defer cancel()

	repo := m.Resource().Repo()
	revision := repo.GetLastRevision(ctx)
	coreNodeName := config.Get().Server.Name

	ctx, cancel1 := context.WithCancel(m.ctx)
	defer cancel1()
	repo.WatchQueue(ctx, revision, func(et dao.EnventType, queue dao.Queue) {
		switch et {
		case dao.PUT:
			// create receiver.
			switch queue.ConsumerType {
			case dao.ConsumerTypeCore:
				if coreNodeName == queue.NodeName {
					log.Info("append queue", zfield.ID(queue.ID))
					// create receiver instance.
					receiver := pubsub.NewPubsub(queue.ID, resource.Metadata{
						Name:       queue.Type.String(),
						Properties: queue.Metadata,
					})

					inboxIns := inbox.New(m.ctx, receiver)
					if _, has := m.inboxes[queue.ID]; has {
						m.inboxes[queue.ID].Close()
					}
					m.inboxes[queue.ID] = inboxIns

					// start consumer inbox.
					log.Info("start consumer inbox", zfield.ID(queue.ID))
					inboxIns.Consume(m.ctx, m.HandleMessage)
					m.containers[inboxIns.ID()] = NewContainer(m.ctx, inboxIns.ID(), m)
				}
			default:
			}
		case dao.DELETE:
			if inboxInst, has := m.inboxes[queue.ID]; has {
				inboxInst.Close()
				m.containers[inboxInst.ID()].Close()
				log.Info("remove inbox", zfield.ID(inboxInst.ID()))
				log.Info("remove queue", zfield.ID(queue.ID),
					zfield.Type(queue.Type.String()), zfield.Name(queue.Name))
			}
		default:
		}
	})
}

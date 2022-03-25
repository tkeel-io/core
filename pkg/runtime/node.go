package runtime

import (
	"context"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/placement"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	xkafka "github.com/tkeel-io/core/pkg/util/kafka"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type NodeConf struct {
	Sources []string
}

type Node struct {
	runtimes        map[string]*Runtime
	dispatch        dispatch.Dispatcher
	resourceManager types.ResourceManager
	mappers         map[string]mapper.Mapper

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewNode(ctx context.Context, resourceManager types.ResourceManager, dispatcher dispatch.Dispatcher) *Node {
	ctx, cacel := context.WithCancel(ctx)
	return &Node{
		ctx:             ctx,
		cancel:          cacel,
		lock:            sync.RWMutex{},
		dispatch:        dispatcher,
		resourceManager: resourceManager,
		runtimes:        make(map[string]*Runtime),
		mappers:         make(map[string]mapper.Mapper),
	}
}

func (n *Node) Start(cfg NodeConf) error {
	log.Info("start node...")

	var elapsed util.ElapsedTime
	n.initializeMetadata()
	for index := range cfg.Sources {
		var err error
		var sourceIns *xkafka.Pubsub
		if sourceIns, err = xkafka.NewKafkaPubsub(cfg.Sources[index]); nil != err {
			return errors.Wrap(err, "create source instance")
		} else if err = sourceIns.Received(n.ctx, n); nil != err {
			return errors.Wrap(err, "consume source")
		}

		rid := sourceIns.ID()
		// create runtime instance.
		log.Info("create runtime instance",
			zfield.ID(rid), zfield.Source(cfg.Sources[index]))

		entityResouce := EntityResource{FlushHandler: n.FlushEntity, RemoveHandler: n.RemoveEntity}
		rt := NewRuntime(n.ctx, entityResouce, rid, n.dispatch, n.resourceManager.Repo())
		for _, mp := range n.mapperSlice() {
			if mc, has := n.mapper(mp)[rt.ID()]; has {
				rt.AppendMapper(*mc)
			}
		}
		n.runtimes[rid] = rt
		placement.Global().Append(placement.Info{ID: sourceIns.ID(), Flag: true})
	}

	log.Debug("start node completed", zfield.Elapsedms(elapsed.ElapsedMilli()))

	return nil
}

func (n *Node) HandleMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	rid := msg.Topic
	if _, has := n.runtimes[rid]; !has {
		log.Error("runtime instance not exists.", zfield.ID(rid),
			zap.Any("header", msg.Headers), zfield.Message(string(msg.Value)))
		return xerrors.ErrRuntimeNotExists
	}

	// load runtime spec.
	rt := n.runtimes[rid]
	rt.DeliveredEvent(context.Background(), msg)
	return nil
}

func (n *Node) initializeMetadata() {
	n.listMetadata()
	go n.watchMetadata()
}

// initialize runtime environments.
func (n *Node) listMetadata() {
	elapsedTime := util.NewElapsed()
	ctx, cancel := context.WithTimeout(n.ctx, 30*time.Second)
	defer cancel()

	repo := n.resourceManager.Repo()
	revision := repo.GetLastRevision(context.Background())
	log.Info("initialize actor manager, mapper loadding...")
	repo.RangeMapper(ctx, revision, func(mappers []dao.Mapper) {
		// 将mapper加入每一个 runtime.
		for _, mp := range mappers {
			// parse mapper.
			mpIns, err := mapper.NewMapper(mp, 1)
			if nil != err {
				log.Error("parse mapper", zap.Error(err),
					zfield.Eid(mp.EntityID), zfield.Mid(mp.ID), zfield.Value(mp))
				continue
			}
			log.Debug("parse mapper", zfield.Eid(mp.EntityID), zfield.Mid(mp.ID))
			n.mappers[mp.ID] = mpIns
		}
	})

	log.Debug("runtime.Environment initialized", zfield.Elapsedms(elapsedTime.ElapsedMilli()))
}

// watchResource watch resources.
func (n *Node) watchMetadata() {
	repo := n.resourceManager.Repo()
	repo.WatchMapper(context.Background(),
		repo.GetLastRevision(context.Background()),
		func(et dao.EnventType, mp dao.Mapper) {
			switch et {
			case dao.DELETE:
				// parse mapper.
				var err error
				var mpIns mapper.Mapper
				log.Info("parse mapper", zfield.Eid(mp.EntityID), zfield.Mid(mp.ID))
				if mpIns, err = mapper.NewMapper(mp, 0); nil != err {
					log.Error("parse mapper", zap.Error(err), zfield.Eid(mp.EntityID), zfield.Mid(mp.ID))
					return
				}

				// remove mapper from all runtime.
				for _, rt := range n.runtimes {
					rt.RemoveMapper(MCache{ID: mpIns.ID()})
				}
			case dao.PUT:
				// parse mapper.
				var err error
				var mpIns mapper.Mapper
				log.Info("parse mapper", zfield.Eid(mp.EntityID), zfield.Mid(mp.ID), zfield.Value(mp))
				if mpIns, err = mapper.NewMapper(mp, 0); nil != err {
					log.Error("parse mapper", zap.Error(err), zfield.Eid(mp.EntityID), zfield.Mid(mp.ID))
					return
				}

				// cache mapper.
				n.mappers[mp.ID] = mpIns
				for rtID, mc := range n.mapper(mpIns) {
					if rt, has := n.runtimes[rtID]; has {
						rt.AppendMapper(*mc)
					}
				}
			}
		})
}

func (n *Node) mapperSlice() []mapper.Mapper {
	mps := []mapper.Mapper{}
	for _, mp := range n.mappers {
		mps = append(mps, mp)
	}
	return mps
}

func (n *Node) mapper(mp mapper.Mapper) map[string]*MCache {
	res := make(map[string]*MCache)
	for eid, tentacles := range mp.Tentacles() {
		// select runtime.
		info := placement.Global().Select(eid)
		if _, exists := res[info.ID]; !exists {
			res[info.ID] = &MCache{
				ID:       mp.ID(),
				Mapper:   mp,
				EntityID: mp.TargetEntity()}
		}
		// append tentacles.
		res[info.ID].Tentacles =
			append(res[info.ID].Tentacles, tentacles...)
	}

	return res
}

func (n *Node) FlushEntity(ctx context.Context, en Entity) error {
	// 1. flush state.
	if err := n.resourceManager.Repo().PutEntity(ctx, en.ID(), en.Raw()); nil != err {
		log.Error("flush entity state storage", zap.Error(err), zfield.Eid(en.ID()))
		return errors.Wrap(err, "flush entity into state storage")
	}

	// 2. flush search engine data.
	indexData := en.Tiled()
	if nil != indexData.Error() {
		log.Error("flush entity search engine, build index data",
			zap.Error(indexData.Error()), zfield.Eid(en.ID()))
		return errors.Wrap(indexData.Error(), "flush entity into search engine, build index data")
	}
	if _, err := n.resourceManager.Search().IndexBytes(ctx, en.ID(), indexData.Raw()); nil != err {
		log.Error("flush entity search engine", zap.Error(err), zfield.Eid(en.ID()))
		return errors.Wrap(err, "flush entity into search engine")
	}

	// 3. flush timeseries data.
	// if _, err := n.resourceManager.TSDB().Write(ctx, &tseries.TSeriesRequest{}); nil != err {
	// 	log.Error("flush entity timeseries database", zap.Error(err), zfield.Eid(en.ID()))
	// }

	return nil
}

func (n *Node) RemoveEntity(ctx context.Context, en Entity) error {
	var err error

	// recover entity state.
	defer func() {
		if nil != err {
			if innerErr := n.FlushEntity(ctx, en); nil != innerErr {
				log.Error("remove entity failed, recover entity state failed", zfield.Eid(en.ID()),
					zfield.Reason(err.Error()), zap.Error(innerErr), zfield.Value(string(en.Raw())))
			}
		}
	}()

	// 1. 从状态存储中删除（可标记）
	if err := n.resourceManager.Repo().
		DelEntity(ctx, en.ID()); nil != err {
		log.Error("remove entity from state storage",
			zap.Error(err), zfield.Eid(en.ID()), zfield.Value(string(en.Raw())))
		return errors.Wrap(err, "remove entity from state storage")
	}

	// 2. 从搜索中删除（可标记）
	if _, err := n.resourceManager.Search().
		DeleteByID(ctx, &v1.DeleteByIDRequest{
			Id:     en.ID(),
			Owner:  en.Owner(),
			Source: en.Source(),
		}); nil != err {
		log.Error("remove entity from state search engine",
			zap.Error(err), zfield.Eid(en.ID()), zfield.Value(string(en.Raw())))
		return errors.Wrap(err, "remove entity from state search engine")
	}

	// 3. 删除etcd中的mapper.
	if err := n.resourceManager.Repo().
		DelMapperByEntity(ctx, &dao.Mapper{
			Owner:    en.Owner(),
			EntityID: en.ID(),
		}); nil != err {
		log.Error("remove entity, remove mapper by entity",
			zap.Error(err), zfield.Eid(en.ID()), zfield.Value(string(en.Raw())))
		return errors.Wrap(err, "remove mapper by entity")
	}
	return nil
}

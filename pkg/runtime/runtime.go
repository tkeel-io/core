package runtime

import (
	"context"
	"fmt"
	"sort"
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
	"github.com/tkeel-io/core/pkg/repository"
	"github.com/tkeel-io/core/pkg/repository/dao"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/core/pkg/util/path"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

type Runtime struct {
	id           string
	node         *Node
	pathTree     *path.Tree
	tentacleTree *path.Tree
	enCache      EntityCache
	entities     map[string]Entity // 存放Runtime的实体.
	dispatcher   dispatch.Dispatcher
	mapperCaches map[string]MCache
	repository   repository.IRepository

	mlock  sync.RWMutex
	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRuntime(ctx context.Context, n *Node, id string, dispatcher dispatch.Dispatcher, repository repository.IRepository) *Runtime {
	ctx, cancel := context.WithCancel(ctx)
	return &Runtime{
		id:           id,
		node:         n,
		enCache:      NewCache(),
		entities:     map[string]Entity{},
		mapperCaches: map[string]MCache{},
		dispatcher:   dispatcher,
		repository:   repository,
		tentacleTree: path.New(),
		pathTree:     path.New(),
		lock:         sync.RWMutex{},
		mlock:        sync.RWMutex{},
		cancel:       cancel,
		ctx:          ctx,
	}
}

func (r *Runtime) ID() string {
	return r.id
}

func (r *Runtime) DeliveredEvent(ctx context.Context, msg *sarama.ConsumerMessage) {
	var err error
	var ev v1.ProtoEvent
	if err = v1.Unmarshal(msg.Value, &ev); nil != err {
		log.Error("decode Event", zap.Error(err))
		return
	}

	r.HandleEvent(ctx, &ev)
}

func (r *Runtime) HandleEvent(ctx context.Context, event v1.Event) error {
	execer, feed := r.PrepareEvent(ctx, event)
	feed = execer.Exec(ctx, feed)

	// call callback once.
	r.handleCallback(ctx, feed)
	return nil
}

func (r *Runtime) PrepareEvent(ctx context.Context, ev v1.Event) (*Execer, *Feed) {
	log.Info("handle event", zfield.ID(ev.ID()), zfield.Eid(ev.Entity()))

	switch ev.Type() {
	case v1.ETSystem:
		execer, feed := r.prepareSystemEvent(ctx, ev)
		execer.postFuncs = append(execer.postFuncs,
			&handlerImpl{fn: r.handlePersistent})
		return execer, feed
	case v1.ETEntity:
		e, _ := ev.(v1.PatchEvent)
		state, err := r.LoadEntity(ev.Entity())
		if nil != err {
			log.Error("load entity", zfield.Eid(ev.Entity()),
				zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
			state = DefaultEntity(ev.Entity())
		}

		execer := &Execer{
			state:    state,
			preFuncs: []Handler{},
			execFunc: state,
			postFuncs: []Handler{
				&handlerImpl{fn: r.handleTentacle},
				&handlerImpl{fn: r.handleComputed},
				&handlerImpl{fn: r.handlePersistent}}}

		return execer, &Feed{
			Err:      err,
			Event:    ev,
			State:    state.Raw(),
			EntityID: ev.Entity(),
			Patches:  conv(e.Patches())}
	case v1.ETCache:
		sender := ev.Attr(v1.MetaSender)
		// load cache.
		state, err := r.enCache.Load(sender)
		if nil != err {
			log.Error("load cache entity", zfield.Header(ev.Attributes()),
				zfield.Eid(ev.Entity()), zfield.ID(ev.ID()), zfield.Sender(sender))
			state = DefaultEntity(sender)
		}

		e, _ := ev.(v1.PatchEvent)
		execer := &Execer{
			state:     state,
			execFunc:  state,
			preFuncs:  []Handler{&handlerImpl{fn: r.handleSubscribe}},
			postFuncs: []Handler{&handlerImpl{fn: r.handleComputed}}}
		return execer, &Feed{
			Event:    ev,
			State:    state.Raw(),
			EntityID: sender,
			Patches:  conv(e.Patches())}
	default:
		return &Execer{}, &Feed{
			Event:    ev,
			State:    DefaultEntity("").Raw(),
			Err:      fmt.Errorf(" unknown RuntimeEvent Type"),
			EntityID: ev.Entity(),
		}
	}
}

// 处理实体生命周期.
func (r *Runtime) prepareSystemEvent(ctx context.Context, event v1.Event) (*Execer, *Feed) {
	log.Info("prepare system event", zfield.ID(event.ID()), zfield.Header(event.Attributes()))
	ev, _ := event.(v1.SystemEvent)
	action := ev.Action()
	operator := action.Operator
	switch v1.SystemOp(operator) {
	case v1.OpCreate:
		log.Info("create entity", zfield.Eid(ev.Entity()),
			zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))

		execer := &Execer{
			state:    DefaultEntity(ev.Entity()),
			preFuncs: []Handler{},
			execFunc: DefaultEntity(ev.Entity()),
			postFuncs: []Handler{
				&handlerImpl{fn: r.handleTentacle},
				&handlerImpl{fn: r.handleComputed},
				&handlerImpl{fn: func(_ context.Context, feed *Feed) *Feed {
					log.Info("create entity successed", zfield.Eid(ev.Entity()),
						zfield.ID(ev.ID()), zfield.Header(ev.Attributes()), zfield.Value(string(action.Data)))
					return feed
				}},
			}}

		// check entity exists.
		if _, exists := r.entities[ev.Entity()]; exists {
			return execer, &Feed{
				Event:    ev,
				EntityID: ev.Entity(),
				Err:      xerrors.ErrEntityAleadyExists}
		}

		// new entity.
		state, err := NewEntity(ev.Entity(), action.GetData())
		if nil != err {
			log.Error("create entity", zfield.Eid(ev.Entity()),
				zfield.Value(string(action.GetData())), zap.Error(err))
			return execer, &Feed{
				Err:      err,
				Event:    ev,
				EntityID: ev.Entity()}
		}

		props := state.Get("properties")
		r.entities[ev.Entity()] = state
		execer.state = state
		execer.execFunc = state
		return execer, &Feed{
			Err:      props.Error(),
			Event:    ev,
			State:    state.Raw(),
			EntityID: ev.Entity(),
			Patches: []Patch{{
				Op:    OpMerge,
				Path:  "properties",
				Value: tdtl.New(props.Raw())}}}
	case v1.OpDelete:
		state, err := r.LoadEntity(ev.Entity())
		if nil != err {
			// TODO: if entity not exists.
			log.Error("delete entity", zfield.Eid(ev.Entity()),
				zfield.Value(string(action.GetData())), zap.Error(err))
			state = DefaultEntity(ev.Entity())
		}
		execer := &Execer{
			state:    state,
			execFunc: state,
			preFuncs: []Handler{
				&handlerImpl{fn: func(ctx context.Context, feed *Feed) *Feed {
					// TODO:
					//		0. 删除etcd中的mapper.
					if err = r.repository.DelMapperByEntity(ctx, &dao.Mapper{
						Owner:    state.Owner(),
						EntityID: state.ID(),
					}); nil != err {
						feed.Err = err
						return feed
					}
					// 1. 从状态存储中删除（可标记）
					if err = r.repository.DelEntity(ctx, state.ID()); nil != err {
						feed.Err = err
						return feed
					}
					// 2. 从搜索中删除（可标记）
					// 3. 从Runtime 中删除.
					delete(r.entities, ev.Entity())
					return feed
				}}},
			postFuncs: []Handler{
				&handlerImpl{fn: r.handleTentacle},
				&handlerImpl{fn: func(_ context.Context, feed *Feed) *Feed {
					log.Info("delete entity successed", zfield.Eid(ev.Entity()),
						zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
					return feed
				}}},
		}

		return execer, &Feed{
			Err:      err,
			Event:    ev,
			State:    state.Raw(),
			EntityID: ev.Entity()}
	default:
		return &Execer{
				state:    DefaultEntity(ev.Entity()),
				preFuncs: []Handler{},
				execFunc: DefaultEntity(ev.Entity()),
				postFuncs: []Handler{
					&handlerImpl{fn: func(_ context.Context, feed *Feed) *Feed {
						log.Error("event type not support", zfield.Eid(ev.Entity()),
							zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
						return feed
					}}}}, &Feed{
				Event:    ev,
				EntityID: ev.Entity(),
				Err:      xerrors.ErrInternal}
	}
}

func (r *Runtime) handleComputed(ctx context.Context, feed *Feed) *Feed {
	// 1. 检查 ret.path 和 订阅列表.
	entityID := feed.EntityID
	mappers := make(map[string]mapper.Mapper)
	for _, change := range feed.Changes {
		for _, node := range r.pathTree.
			MatchPrefix(path.FmtWatchKey(entityID, change.Path)) {
			tentacle, _ := node.(mapper.Tentacler)
			if tentacle.Type() == mapper.TentacleTypeMapper {
				mappers[tentacle.TargetID()] = tentacle.Mapper()
			}
		}
	}

	patches := make(map[string][]*v1.PatchData)
	for id, mp := range mappers {
		target := mp.TargetEntity()
		log.Debug("compute mapper",
			zfield.Eid(entityID), zfield.Mid(id))
		result := r.computeMapper(ctx, mp)
		for path, val := range result {
			patches[target] = append(
				patches[target],
				&v1.PatchData{
					Operator: string(OpReplace),
					Path:     path,
					Value:    val.Raw(),
				})
		}
	}

	// 2. dispatch.send()
	for target, patch := range patches {
		r.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
			Id:        util.IG().EvID(),
			Timestamp: time.Now().UnixNano(),
			Metadata: map[string]string{
				v1.MetaType:     string(v1.ETEntity),
				v1.MetaEntityID: target},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: patch,
				}},
		})
	}

	return feed
}

func (r *Runtime) computeMapper(ctx context.Context, mp mapper.Mapper) map[string]tdtl.Node {
	var (
		has bool
		err error
		mc  MCache
	)

	r.mlock.RLock()
	if mc, has = r.mapperCaches[mp.ID()]; !has {
		r.mlock.RUnlock()
		return map[string]tdtl.Node{}
	}
	r.mlock.RUnlock()

	// construct mapper input.
	in := make(map[string]tdtl.Node)
	for _, tentacle := range mc.Tentacles {
		for _, item := range tentacle.Items() {
			var state Entity
			// get value from entities.
			if state, has = r.entities[item.EntityID]; has {
				in[item.String()] = state.Get(item.PropertyKey)
				continue
			}
			// get value from cache.
			if state, err = r.enCache.Load(item.EntityID); nil == err {
				in[item.String()] = state.Get(item.PropertyKey)
			}
		}
	}

	// ignore empty input.
	if len(in) == 0 {
		log.Warn("ignore empty input", zfield.Mid(mp.ID()))
		return map[string]tdtl.Node{}
	}

	var out map[string]tdtl.Node
	if out, err = mp.Exec(in); nil != err {
		log.Error("exec mapper", zfield.ID(mp.ID()), zfield.Eid(mp.TargetEntity()))
		return map[string]tdtl.Node{}
	}

	// clean nil feed.
	for path, val := range out {
		if val == nil || val.Type() == tdtl.Null ||
			val.Type() == tdtl.Undefined || val.Error() != nil {
			log.Warn("invalid computed feed", zap.Any("value", val), zfield.Mid(mp.ID()))
			delete(out, path)
		}
	}

	return out
}

func (r *Runtime) handleTentacle(ctx context.Context, feed *Feed) *Feed {
	// 1. 检查 ret.path 和 订阅列表.
	var targets sort.StringSlice
	entityID := feed.EntityID
	var patches = make(map[string][]*v1.PatchData)
	for _, change := range feed.Changes {
		for _, node := range r.pathTree.
			MatchPrefix(path.FmtWatchKey(entityID, change.Path)) {
			tentacle, _ := node.(mapper.Tentacler)
			if entityID != tentacle.TargetID() &&
				mapper.TentacleTypeEntity == tentacle.Type() {
				targets = append(targets, tentacle.TargetID())
			}
		}

		targets = util.Unique(targets)
		for _, target := range targets {
			patches[target] = append(
				patches[target],
				&v1.PatchData{
					Path:     change.Path,
					Operator: string(OpReplace),
					Value:    change.Value.Raw(),
				})
		}

		// clean targets.
		targets = []string{}
	}

	// 2. dispatch.send()
	for target, patch := range patches {
		// check target entity placement.
		info := placement.Global().Select(target)
		if info.ID == r.id {
			log.Debug("target entity belong this runtime, ignore dispatch.",
				zfield.Sender(entityID), zfield.Eid(target), zfield.ID(info.ID))
			continue
		}

		// dispatch cache event.
		r.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
			Id:        util.IG().EvID(),
			Timestamp: time.Now().UnixNano(),
			Metadata: map[string]string{
				v1.MetaType:     string(v1.ETCache),
				v1.MetaEntityID: target,
				v1.MetaSender:   entityID},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: patch,
				}},
		})
	}

	return feed
}

func (r *Runtime) handleCallback(ctx context.Context, feed *Feed) error {
	var err error
	event := feed.Event
	log.Debug("handle event, callback.", zfield.ID(event.ID()),
		zfield.Eid(event.Entity()), zfield.Header(event.Attributes()))

	if event.CallbackAddr() != "" {
		if feed.Err == nil {
			// 需要注意的是：为了精炼逻辑，runtime内部只是对api返回变更后实体的最新状态，而不做API结果的组装.
			ev := &v1.ProtoEvent{
				Id:        event.ID(),
				Timestamp: time.Now().UnixNano(),
				Callback:  event.CallbackAddr(),
				Metadata:  event.Attributes(),
				Data: &v1.ProtoEvent_RawData{
					RawData: feed.State}}
			ev.SetType(v1.ETCallback)
			ev.SetAttr(v1.MetaResponseStatus, string(types.StatusOK))
			err = r.dispatcher.Dispatch(ctx, ev)
		} else {
			ev := &v1.ProtoEvent{
				Id:        event.ID(),
				Timestamp: time.Now().UnixNano(),
				Callback:  event.CallbackAddr(),
				Metadata:  event.Attributes(),
				Data: &v1.ProtoEvent_RawData{
					RawData: []byte(`{}`)}}
			ev.SetType(v1.ETCallback)
			ev.SetAttr(v1.MetaResponseStatus, string(types.StatusError))
			ev.SetAttr(v1.MetaResponseErrCode, feed.Err.Error())
			err = r.dispatcher.Dispatch(ctx, ev)
		}
	}

	if nil != err {
		log.Error("handle event, callback.", zfield.ID(event.ID()),
			zap.Error(err), zfield.Eid(event.Entity()), zfield.Header(event.Attributes()))
	}

	return errors.Wrap(err, "handle callback")
}

func (r *Runtime) handlePersistent(ctx context.Context, feed *Feed) *Feed {
	en, ok := r.entities[feed.EntityID]
	if !ok {
		// entity has been deleted.
		return feed
	}
	r.node.FlushEntity(ctx, en)
	return feed
}

func (r *Runtime) AppendMapper(mc MCache) {
	r.mlock.Lock()
	// remove if existed.
	if _, exists := r.mapperCaches[mc.ID]; exists {
		mc0 := r.mapperCaches[mc.ID]
		delete(r.mapperCaches, mc.ID)
		for _, tentacle := range mc0.Tentacles {
			for _, item := range tentacle.Items() {
				r.pathTree.Remove(item.String(), tentacle)
			}
		}
	}

	r.mapperCaches[mc.ID] = mc
	for _, tentacle := range mc.Tentacles {
		for _, item := range tentacle.Items() {
			r.pathTree.Add(item.String(), tentacle)
		}
	}
	r.mlock.Unlock()

	// initialize mapper, exec mapper once.
	r.initializeMapper(context.TODO(), mc)

	r.pathTree.Print()
}

func (r *Runtime) initializeMapper(ctx context.Context, mc MCache) {
	if mapper.VersionInited != mc.Mapper.Version() {
		return
	}

	var items []mapper.WatchKey
	for _, tentacle := range mc.Tentacles {
		switch tentacle.Type() {
		case mapper.TentacleTypeEntity:
			items = append(items, tentacle.Items()...)
		default:
		}
	}

	feeds := make(map[string]*Feed)
	for _, item := range items {
		var err error
		var val tdtl.Node
		var state Entity
		if state, err = r.LoadEntity(item.EntityID); nil != err {
			log.Warn("load entity", zap.Error(err), zfield.Eid(item.EntityID))
			continue
		} else if val = state.Get(item.PropertyKey); nil != val.Error() {
			log.Warn("get entity property", zap.Error(val.Error()), zfield.Eid(item.EntityID))
			continue
		}

		if _, has := feeds[item.EntityID]; !has {
			feeds[item.EntityID] = &Feed{
				EntityID: item.EntityID,
			}
		}

		feeds[item.EntityID].Changes =
			append(feeds[item.EntityID].Changes,
				Patch{
					Op:    OpReplace,
					Path:  item.PropertyKey,
					Value: tdtl.New(val.Raw())})
	}

	// handle subscribe.
	for _, feed := range feeds {
		r.handleTentacle(ctx, feed)
	}
}

func (r *Runtime) RemoveMapper(mc MCache) {
	r.mlock.Lock()
	defer r.mlock.Unlock()
	if _, exists := r.mapperCaches[mc.ID]; !exists {
		return
	}

	mc = r.mapperCaches[mc.ID]
	delete(r.mapperCaches, mc.ID)
	for _, tentacle := range mc.Tentacles {
		for _, item := range tentacle.Items() {
			r.pathTree.Remove(item.String(), tentacle)
		}
	}

	r.pathTree.Print()
}

func (r *Runtime) LoadEntity(id string) (Entity, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if state, ok := r.entities[id]; ok {
		return state, nil
	}

	// load from state storage.
	jsonData, err := r.repository.GetEntity(context.TODO(), id)
	if nil != err {
		log.Warn("load entity from state storage",
			zfield.Eid(id), zfield.Reason(err.Error()))
		return nil, errors.Wrap(err, "load entity")
	}

	// create entity instance.
	en, err := NewEntity(id, jsonData)
	if nil != err {
		log.Warn("create entity instance",
			zfield.Eid(id), zfield.Reason(err.Error()))
		return nil, errors.Wrap(err, "create entity instance")
	}

	return en, nil
}

func conv(patches []*v1.PatchData) []Patch {
	res := make([]Patch, 0)
	for _, patch := range patches {
		res = append(res, Patch{
			Op:    PatchOp(patch.Operator),
			Path:  patch.Path,
			Value: tdtl.New(patch.Value),
		})
	}
	return res
}

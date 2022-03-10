package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/mapper"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/core/pkg/util"
	"github.com/tkeel-io/core/pkg/util/path"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

type Runtime struct {
	id           string
	pathTree     *path.Tree
	caches       map[string]Entity // 存放其他Runtime的实体.
	entities     map[string]Entity // 存放Runtime的实体.
	dispatcher   dispatch.Dispatcher
	mapperCaches map[string]MCache

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRuntime(ctx context.Context, id string, dispatcher dispatch.Dispatcher) *Runtime {
	ctx, cancel := context.WithCancel(ctx)
	return &Runtime{
		id:           id,
		caches:       map[string]Entity{},
		entities:     map[string]Entity{},
		mapperCaches: map[string]MCache{},
		dispatcher:   dispatcher,
		pathTree:     path.New(),
		lock:         sync.RWMutex{},
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
	execer, result := r.PrepareEvent(ctx, event)
	result = execer.Exec(ctx, result)

	// call callback once.
	r.handleCallback(ctx, result)
	return result.Err
}

func (r *Runtime) PrepareEvent(ctx context.Context, ev v1.Event) (*Execer, *Result) {
	log.Info("handle event", zfield.ID(ev.ID()), zfield.Eid(ev.Entity()))

	switch ev.Type() {
	case v1.ETSystem:
		execer, result := r.handleSystemEvent(ctx, ev)
		return execer, result
	case v1.ETEntity:
		e, _ := ev.(v1.PatchEvent)
		state, err := r.LoadEntity(ev.Entity())
		if nil != err {
			log.Error("load entity", zfield.Eid(ev.Entity()),
				zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
			state = DefaultEntity(ev.Entity())
		}

		return &Execer{
				state:    state,
				preFuncs: []Handler{},
				execFunc: state,
				postFuncs: []Handler{
					&handlerImpl{fn: r.handleSubscribe},
					&handlerImpl{fn: r.handleComputed},
				},
			}, &Result{
				Err:      err,
				Event:    ev,
				State:    state.Raw(),
				EntityID: ev.Entity(),
				Patches:  conv(e.Patches())}
	case v1.ETCache:
		return &Execer{
			state:    nil,
			preFuncs: []Handler{},
			postFuncs: []Handler{
				&handlerImpl{fn: r.handleComputed},
			},
		}, &Result{}
	default:
		return &Execer{
				state:     nil,
				preFuncs:  []Handler{},
				postFuncs: []Handler{},
			}, &Result{
				Err:      fmt.Errorf(" unknown RuntimeEvent Type"),
				EntityID: ev.Entity(),
			}
	}
}

// 处理实体生命周期.
func (r *Runtime) handleSystemEvent(ctx context.Context, event v1.Event) (*Execer, *Result) {
	log.Info("handle system event", zfield.ID(event.ID()), zfield.Header(event.Attributes()))
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
				&handlerImpl{fn: r.handleSubscribe},
				&handlerImpl{fn: r.handleComputed},
				&handlerImpl{fn: func(ctx context.Context, result *Result) *Result {
					log.Info("create entity successed", zfield.Eid(ev.Entity()),
						zfield.ID(ev.ID()), zfield.Header(ev.Attributes()), zfield.Value(string(action.Data)))
					return result
				}},
			}}

		// check entity exists.
		if _, exists := r.entities[ev.Entity()]; exists {
			return execer, &Result{
				Event:    ev,
				EntityID: ev.Entity(),
				Err:      xerrors.ErrEntityAleadyExists,
				State:    DefaultEntity(ev.Entity()).Raw()}
		}

		// new entity.
		state, err := NewEntity(ev.Entity(), action.GetData())
		if nil != err {
			log.Error("create entity", zfield.Eid(ev.Entity()),
				zfield.Value(string(action.GetData())), zap.Error(err))
			return execer, &Result{
				Err:      err,
				Event:    ev,
				EntityID: ev.Entity(),
				State:    DefaultEntity(ev.Entity()).Raw()}
		}

		r.entities[ev.Entity()] = state
		execer.state = state
		execer.execFunc = state
		return execer, &Result{
			Event:    ev,
			EntityID: ev.Entity(),
			State:    action.GetData()}
	case v1.OpDelete:
		state, err := r.LoadEntity(ev.Entity())
		return &Execer{
				state:    state,
				execFunc: state,
				preFuncs: []Handler{
					&handlerImpl{fn: func(ctx context.Context, result *Result) *Result {
						// TODO:
						//		0. 删除etcd中的mapper.
						//		1. 从状态存储中删除（可标记）
						//		2. 从搜索中删除（可标记）
						// 		3. 从Runtime 中删除.
						delete(r.entities, ev.Entity())
						return result
					}}},
				postFuncs: []Handler{
					&handlerImpl{fn: r.handleSubscribe},
					&handlerImpl{fn: func(ctx context.Context, result *Result) *Result {
						log.Info("delete entity successed", zfield.Eid(ev.Entity()),
							zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
						return result
					}}},
			}, &Result{
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
					&handlerImpl{fn: func(ctx context.Context, result *Result) *Result {
						log.Error("event type not support", zfield.Eid(ev.Entity()),
							zfield.ID(ev.ID()), zfield.Header(ev.Attributes()))
						return result
					}}}}, &Result{
				Event:    ev,
				EntityID: ev.Entity(),
				Err:      xerrors.ErrInternal}
	}
}

func (r *Runtime) handleComputed(ctx context.Context, result *Result) *Result {
	// 1. 检查 ret.path 和 订阅列表.
	entityID := result.EntityID
	mappers := make(map[string]mapper.Mapper)
	for _, change := range result.Patches {
		for _, node := range r.pathTree.
			MatchPrefix(entityID + change.Path) {
			tentacle, _ := node.(mapper.Tentacler)
			if tentacle.Type() == "mapper" {
				mappers[tentacle.TargetID()] = tentacle.Mapper()
			}
		}
	}

	patches := make(map[string]Patch)
	for id, mp := range mappers {
		log.Debug("compute mapper",
			zfield.Eid(entityID), zfield.Mid(id))
		for path, val := range r.computeMapper(ctx, mp) {
			patches[path] = Patch{
				Op:    OpReplace,
				Path:  path,
				Value: tdtl.New(val.Raw()),
			}
		}
	}

	for _, patch := range patches {
		result.Patches = append(result.Patches, patch)
	}

	return result
}

func (r *Runtime) computeMapper(ctx context.Context, mp mapper.Mapper) map[string]tdtl.Node {
	in := make(map[string]tdtl.Node)

	// construct mapper input.

	out, err := mp.Exec(in)
	if nil != err {
		log.Error("exec mapper",
			zfield.ID(mp.ID()),
			zfield.Eid(mp.TargetEntity()))
		return map[string]tdtl.Node{}
	}

	return out
}

func (r *Runtime) handleSubscribe(ctx context.Context, result *Result) *Result {
	// 1. 检查 ret.path 和 订阅列表.
	var targets []string
	entityID := result.EntityID
	var patches = make(map[string][]*v1.PatchData)
	for _, change := range result.Patches {
		for _, node := range r.pathTree.MatchPrefix(entityID + change.Path) {
			mp, _ := node.(mapper.Tentacler)
			if entityID != mp.TargetID() {
				targets = append(targets, mp.TargetID())
			}
		}

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
		r.dispatcher.Dispatch(ctx, &v1.ProtoEvent{
			Id:        util.IG().EvID(),
			Timestamp: time.Now().UnixNano(),
			Metadata: map[string]string{
				v1.MetaType:     string(v1.ETCache),
				v1.MetaEntityID: entityID,
				v1.MetaSender:   target},
			Data: &v1.ProtoEvent_Patches{
				Patches: &v1.PatchDatas{
					Patches: patch,
				}},
		})
	}

	return result
}

func (r *Runtime) handleCallback(ctx context.Context, result *Result) error {
	var err error
	event := result.Event
	log.Debug("handle event, callback.", zfield.ID(event.ID()),
		zfield.Eid(event.Entity()), zfield.Header(event.Attributes()))

	if event.CallbackAddr() != "" {
		if result.Err == nil {
			// 需要注意的是：为了精炼逻辑，runtime内部只是对api返回变更后实体的最新状态，而不做API结果的组装.
			ev := &v1.ProtoEvent{
				Id:        event.ID(),
				Timestamp: time.Now().UnixNano(),
				Callback:  event.CallbackAddr(),
				Metadata:  event.Attributes(),
				Data: &v1.ProtoEvent_RawData{
					RawData: result.State}}
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
					RawData: []byte{}}}
			ev.SetType(v1.ETCallback)
			ev.SetAttr(v1.MetaResponseStatus, string(types.StatusError))
			ev.SetAttr(v1.MetaResponseErrCode, result.Err.Error())
			err = r.dispatcher.Dispatch(ctx, ev)
		}
	}

	if nil != err {
		log.Error("handle event, callback.", zfield.ID(event.ID()),
			zap.Error(err), zfield.Eid(event.Entity()), zfield.Header(event.Attributes()))
	}

	return errors.Wrap(err, "handle callback")
}

func (r *Runtime) AppendMapper(mc MCache) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.mapperCaches[mc.ID] = mc
	for _, tantacle := range mc.Tentacles {
		for _, item := range tantacle.Items() {
			r.pathTree.Add(item.String(), tantacle)
		}
	}

	r.pathTree.Print()
}

func (r *Runtime) RemoveMapper(mc MCache) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, exists := r.mapperCaches[mc.ID]; !exists {
		return
	}

	mc = r.mapperCaches[mc.ID]
	delete(r.mapperCaches, mc.ID)
	for _, tantacle := range mc.Tentacles {
		for _, item := range tantacle.Items() {
			r.pathTree.Remove(item.String(), tantacle)
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

	return nil, xerrors.ErrEntityNotFound
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

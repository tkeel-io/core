package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/core/pkg/types"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tdtl"
	"go.uber.org/zap"
)

type Runtime struct {
	id         string
	caches     map[string]Entity //存放其他Runtime的实体
	entities   map[string]Entity //存放Runtime的实体
	dispatcher dispatch.Dispatcher
	//inbox    Inbox

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRuntime(ctx context.Context, id string, dispatcher dispatch.Dispatcher) *Runtime {
	ctx, cancel := context.WithCancel(ctx)
	return &Runtime{
		id:         id,
		caches:     map[string]Entity{},
		entities:   map[string]Entity{},
		dispatcher: dispatcher,
		lock:       sync.RWMutex{},
		cancel:     cancel,
		ctx:        ctx,
	}
}

func (e *Runtime) DeliveredEvent(ctx context.Context, msg *sarama.ConsumerMessage) {
	var err error
	var ev v1.ProtoEvent
	if err = v1.Unmarshal(msg.Value, &ev); nil != err {
		log.Error("decode Event", zap.Error(err))
		return
	}

	e.HandleEvent(ctx, &ev)
}

func (e *Runtime) HandleEvent(ctx context.Context, event v1.Event) error {
	execer, result := e.PrepareEvent(ctx, event)
	result = execer.Exec(ctx, result)
	return result.Err
}

func (r *Runtime) PrepareEvent(ctx context.Context, ev v1.Event) (*Execer, *Result) {
	log.Info("handle event", zfield.ID(ev.ID()), zfield.Eid(ev.Entity()))
	//2.1 实体必须包含 entityID，创建、删除等消息：由 Runtime 处理
	//    实体配置重载？Mapper变化了（Mapper包括 订阅-source、执行-target）

	switch ev.Type() {
	case v1.ETSystem:
		process, state, result :=
			r.handleSystemEvent(ctx, ev)

		return &Execer{
			state:     state,
			preFuncs:  []Handler{},
			execFunc:  process,
			postFuncs: []Handler{},
		}, result
	case v1.ETEntity:
		e, _ := ev.(v1.PatchEvent)
		state, err := r.LoadEntity(ev.Entity())
		return &Execer{
				state:     state,
				preFuncs:  []Handler{},
				execFunc:  state,
				postFuncs: []Handler{},
			}, &Result{
				Err:     err,
				event:   ev,
				State:   state.Raw(),
				Patches: conv(e.Patches())}
	case v1.ETCache:
		return &Execer{
			state:     nil,
			preFuncs:  []Handler{},
			postFuncs: []Handler{},
		}, &Result{}
	default:
		return &Execer{
			state:     nil,
			preFuncs:  []Handler{},
			postFuncs: []Handler{},
		}, &Result{Err: fmt.Errorf(" unknown RuntimeEvent Type")}
	}
}

//处理实体生命周期
func (r *Runtime) handleSystemEvent(ctx context.Context, event v1.Event) (Handler, Entity, *Result) {
	log.Info("handle system event", zfield.ID(event.ID()), zfield.Header(event.Attributes()))
	ev, _ := event.(v1.SystemEvent)
	action := ev.Action()
	operator := action.Operator
	switch v1.SystemOp(operator) {
	case v1.OpCreate:
		// create entity.
		state, err := NewEntity(ev.Entity(), action.GetData())
		if nil != err {
			log.Error("create entity", zfield.Eid(ev.Entity()),
				zfield.Value(string(action.GetData())), zap.Error(err))
		}

		// return process state handler.
		return &handlerImpl{fn: func(ctx context.Context, result *Result) *Result {
				if nil != result.Err {
					return result
				}

				log.Info("create entity", zfield.Eid(ev.Entity()),
					zfield.ID(event.ID()), zfield.Header(event.Attributes()))

				// check entity exists.
				if _, exists := r.entities[ev.Entity()]; exists {
					return &Result{Err: xerrors.ErrEntityAleadyExists}
				}

				// new Entity instance.
				enIns, innerErr := NewEntity(ev.Entity(), action.GetData())
				if nil != err {
					log.Error("create entity", zfield.Eid(ev.Entity()),
						zfield.Value(string(action.GetData())), zap.Error(innerErr))
				}

				// TODO: parse get changes.

				// store entity.
				r.entities[ev.Entity()] = enIns
				return &Result{State: enIns.Raw(), Err: innerErr}
			}}, state, &Result{
				Err:   err,
				State: action.GetData(),
				event: ev}
	case v1.OpDelete:
		state, err := r.LoadEntity(ev.Entity())
		return &handlerImpl{fn: func(ctx context.Context, result *Result) *Result {
				// TODO:
				//		0. 删除etcd中的mapper.
				//		1. 从状态存储中删除（可标记）
				//		2. 从搜索中删除（可标记）
				// 		3. 从Runtime 中删除.
				delete(r.entities, ev.Entity())
				return result
			}}, state, &Result{
				Err:   err,
				event: ev,
				State: state.Raw()}
	default:
		return &handlerImpl{fn: func(ctx context.Context, result *Result) *Result {
				// do nothing...
				return result
			}}, DefaultEntity(event.Entity()),
			&Result{Err: xerrors.ErrInternal}
	}
}

//Runtime 处理 event
func (e *Runtime) Process(ctx context.Context, event v1.Event) (*Result, error) {
	panic("implement me")
	//2.2  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
	//              mappers[entityID]=Mapper(dispatch,stateBytes)
	//3.3 触发对应的Mapper（执行-target） -> Patch
	//3.4 更新实体（target），记录下变更   -> Result(a.p2)
	//EntityID := event.ID
	//mapper, err := e.Mapper(EntityID)
	//if err != nil {
	//	return nil, err
	//}
	//ret, err := mapper.Handle(ctx, event)
	//return nil, nil
}

//处理订阅
func (e *Runtime) handleSubscribe(ctx context.Context, ret *Result) *Result {
	if nil != ret.Err {
		return ret
	}

	//@TODO
	// 1. 检查 ret.path 和 订阅列表
	// 2. 执行对应的订阅，
	// 3. dispatch.send()
	return ret
}

func (r *Runtime) handleCallback(ctx context.Context, event v1.Event, ret *Result) *Result {
	if event.CallbackAddr() != "" {
		switch ret.Err {
		case nil:
			// 需要注意的是：为了精炼逻辑，runtime内部只是对api返回变更后实体的最新状态，而不做API结果的组装.
			ev := &v1.ProtoEvent{
				Id:        event.ID(),
				Timestamp: time.Now().UnixNano(),
				Callback:  event.CallbackAddr(),
				Metadata:  event.Attributes(),
				Data: &v1.ProtoEvent_RawData{
					RawData: ret.State,
				},
			}
			ev.SetType(v1.ETCallback)
			ev.SetAttr(v1.MetaResponseStatus, string(types.StatusOK))
			r.dispatcher.Dispatch(ctx, ev)
		default:
			ev := &v1.ProtoEvent{
				Id:        event.ID(),
				Timestamp: time.Now().UnixNano(),
				Callback:  event.CallbackAddr(),
				Metadata:  event.Attributes(),
				Data: &v1.ProtoEvent_RawData{
					RawData: []byte{},
				},
			}
			ev.SetType(v1.ETCallback)
			ev.SetAttr(v1.MetaResponseStatus, string(types.StatusError))
			ev.SetAttr(v1.MetaResponseErrCode, ret.Err.Error())
			r.dispatcher.Dispatch(ctx, ev)
		}
	}

	return ret
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
	var res []Patch
	for _, patch := range patches {
		res = append(res, Patch{
			Op:    PatchOp(patch.Operator),
			Path:  patch.Path,
			Value: tdtl.New(patch.Value),
		})
	}
	return res
}

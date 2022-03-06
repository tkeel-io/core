package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/core/pkg/dispatch"
	xerrors "github.com/tkeel-io/core/pkg/errors"
	zfield "github.com/tkeel-io/core/pkg/logger"
	"github.com/tkeel-io/kit/log"
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

func NewRuntime(ctx context.Context, id string) *Runtime {
	ctx, cancel := context.WithCancel(ctx)
	return &Runtime{
		id:       id,
		caches:   map[string]Entity{},
		entities: map[string]Entity{},
		lock:     sync.RWMutex{},
		cancel:   cancel,
		ctx:      ctx,
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
	var ret *Result
	//1.BeforeProcess
	//1. 升级执行的环境
	//1.1 处理 Entity 的创建、删除
	//1.2 Cache 消息直接更新 Runtime 的 caches
	//1.3 更新实体，记录下变更    -> Result
	ret = e.UpdateWithEvent(ctx, event)

	//2.Process
	//2.1  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
	//              mappers[entityID]=Mapper(dispatch,stateBytes)
	//2.2 触发对应的Mapper（执行-target） -> Patch
	//2.3 更新实体（target），记录下变更   -> Result(a.p2)
	//@TODO Process
	//ret, err = e.Process(ctx, event)

	//3.AfterProcess
	//3.1 依照订阅发布实体变更  handleSubscribe
	ret = e.handleSubscribe(ctx, ret)
	//4.2 处理 API 回调       handleCallback
	ret = e.handleCallback(ctx, event, ret)
	return ret.Err
}

func (r *Runtime) UpdateWithEvent(ctx context.Context, event v1.Event) *Result {
	//2.1 实体必须包含 entityID，创建、删除等消息：由 Runtime 处理
	//    实体配置重载？Mapper变化了（Mapper包括 订阅-source、执行-target）
	switch event.Type() {
	case v1.ETSystem:
		return r.handleSystemEvent(ctx, event)
	case v1.ETEntity:
		EntityID := event.Entity()
		entity, err := r.LoadEntity(EntityID)
		if err != nil {
			return &Result{Err: err}
		}
		return entity.Handle(ctx, event)
	case v1.ETCache:
		return r.handleCacheEvent(ctx, event)
	default:
		return &Result{Err: fmt.Errorf(" unknown RuntimeEvent Type")}
	}
}

//处理实体生命周期
func (r *Runtime) handleSystemEvent(ctx context.Context, event v1.Event) *Result {
	ev, _ := event.(v1.SystemEvent)
	action := ev.Action()
	operator := action.Operator
	switch v1.SystemOp(operator) {
	case v1.OpCreate:
		// check entity exists.
		if _, exists := r.entities[ev.Entity()]; exists {
			return &Result{Err: xerrors.ErrEntityAleadyExists}
		}

		// create entity.
		en, err := NewEntity(ev.Entity(), action.GetData())
		if nil != err {
			log.Error("create entity", zfield.Eid(ev.Entity()))
			return &Result{Err: err}
		}

		// TODO: parse get changes.

		// store entity.
		r.entities[ev.Entity()] = en
		return &Result{State: en.Raw()}
	case v1.OpDelete:
		// TODO:
		//		1. 从状态存储中删除（可标记）
		//		2. 从搜索中删除（可标记）
		// 		3. 从Runtime 中删除.
		delete(r.entities, ev.Entity())
	default:
		return &Result{Err: xerrors.ErrInternal}
	}
	return &Result{Err: xerrors.ErrInternal}
}

//处理Cache
func (e *Runtime) handleCacheEvent(ctx context.Context, event v1.Event) *Result {
	panic("implement me")
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
	//@TODO
	// 1. 检查 ret.path 和 订阅列表
	// 2. 执行对应的订阅，
	// 3. dispatch.send()
	return ret
}

func (e *Runtime) handleCallback(ctx context.Context, event v1.Event, ret *Result) *Result {
	cbAddr := event.CallbackAddr()
	if cbAddr != "" {
	}
	return &Result{Err: ret.Err}
}

func (e *Runtime) LoadEntity(id string) (Entity, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if state, ok := e.entities[id]; ok {
		return state, nil
	}

	return nil, nil
}

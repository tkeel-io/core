package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	v1 "github.com/tkeel-io/core/api/core/v1"
	"github.com/tkeel-io/kit/log"
	"go.uber.org/zap"
)

type Runtime struct {
	id       string
	caches   map[string]Entity //存放其他Runtime的实体
	entities map[string]Entity //存放Runtime的实体
	dispatch Dispatcher
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

func (e *Runtime) HandleEvent(ctx context.Context, event v1.Event) (*Result, error) {
	var (
		ret *Result
		err error
	)

	//1.BeforeProcess
	//1. 升级执行的环境
	//1.1 处理 Entity 的创建、删除
	//1.2 Cache 消息直接更新 Runtime 的 caches
	//1.3 更新实体，记录下变更    -> Result
	ret, err = e.UpdateWithEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	//2.Process
	//2.1  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
	//              mappers[entityID]=Mapper(dispatch,stateBytes)
	//2.2 触发对应的Mapper（执行-target） -> Patch
	//2.3 更新实体（target），记录下变更   -> Result(a.p2)
	//@TODO Process
	//ret, err = e.Process(ctx, event)

	//3.AfterProcess
	//3.1 依照订阅发布实体变更  handleSubscribe
	_, err = e.handleSubscribe(ctx, ret)
	if err != nil {
		return nil, err
	}
	//4.2 处理 API 回调       handleCallback
	_, err = e.handleCallback(ctx, event, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (e *Runtime) UpdateWithEvent(ctx context.Context, event v1.Event) (*Result, error) {
	//2.1 实体必须包含 entityID，创建、删除等消息：由 Runtime 处理
	//    实体配置重载？Mapper变化了（Mapper包括 订阅-source、执行-target）
	switch EventType(event.Type()) {
	case ETRuntime:
		return e.handleRuntimeEvent(ctx, event)
	case ETEntity:
		EntityID := event.Entity()
		entity, err := e.Entity(EntityID)
		if err != nil {
			return nil, err
		}
		ret, err := entity.Handle(ctx, event)
		return ret, err
	case ETCache:
		return e.handleCacheEvent(ctx, event)
	default:
		return nil, fmt.Errorf(" unknown RuntimeEvent Type")
	}
}

//处理实体生命周期
func (e *Runtime) handleRuntimeEvent(ctx context.Context, event v1.Event) (*Result, error) {
	panic("implement me")
}

//处理Cache
func (e *Runtime) handleCacheEvent(ctx context.Context, event v1.Event) (*Result, error) {
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
func (e *Runtime) handleSubscribe(ctx context.Context, ret *Result) (*Result, error) {
	//@TODO
	// 1. 检查 ret.path 和 订阅列表
	// 2. 执行对应的订阅，
	// 3. dispatch.send()
	return nil, nil
}

func (e *Runtime) handleCallback(ctx context.Context, event v1.Event, ret *Result) (*Result, error) {
	//@TODO 处理回调
	// 1. dispatch.respose(ret)
	return nil, nil
}

func (e *Runtime) Entity(id string) (Entity, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if state, ok := e.entities[id]; ok {
		return state, nil
	}

	return nil, nil
}

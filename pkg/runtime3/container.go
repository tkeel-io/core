package runtime3

import (
	"context"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
)

type Container struct {
	id       string
	caches   map[string]Entity //存放其他Container的实体
	entities map[string]Entity //存放Container的实体
	dispatch Dispatcher
	//inbox    Inbox

	lock   sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewContainer(ctx context.Context, id string) *Container {
	ctx, cancel := context.WithCancel(ctx)
	return &Container{
		id:       id,
		caches:   map[string]Entity{},
		entities: map[string]Entity{},
		lock:     sync.RWMutex{},
		cancel:   cancel,
		ctx:      ctx,
	}
}

func (e *Container) DeliveredEvent(ctx context.Context, event interface{}) {
	// 1. 通过 inbox 实现event 转换. 暂时忽略Inbox.
	msg, _ := event.(*sarama.ConsumerMessage)

	ev := deliveredEvent(msg)
	e.HandleEvent(ctx, ev)
}

func (e *Container) HandleEvent(ctx context.Context, event *ContainerEvent) (*Result, error) {
	var (
		ret *Result
		err error
	)

	//1.BeforeProcess
	//1. 升级执行的环境
	//1.1 处理 Entity 的创建、删除
	//1.2 Cache 消息直接更新 Container 的 caches
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

func (e *Container) UpdateWithEvent(ctx context.Context, event *ContainerEvent) (*Result, error) {
	//2.1 实体必须包含 entityID，创建、删除等消息：由 Container 处理
	//    实体配置重载？Mapper变化了（Mapper包括 订阅-source、执行-target）
	switch event.Type {
	case OpContainer:
		return e.handleMangerEvent(ctx, event)
	case OpEntity:
		EntityID := event.ID
		entity, err := e.Entity(EntityID)
		if err != nil {
			return nil, err
		}
		ret, err := entity.Handle(ctx, event.Value)
		return ret, err
	case OpCache:
		return e.handleCacheEvent(ctx, event)
	default:
		return nil, fmt.Errorf(" unknown ContainerEvent Type")
	}
}

//处理实体生命周期
func (e *Container) handleMangerEvent(ctx context.Context, event *ContainerEvent) (*Result, error) {
	panic("implement me")
}

//处理Cache
func (e *Container) handleCacheEvent(ctx context.Context, event *ContainerEvent) (*Result, error) {
	panic("implement me")
}

//Container 处理 event
func (e *Container) Process(ctx context.Context, event *ContainerEvent) (*Result, error) {
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
func (e *Container) handleSubscribe(ctx context.Context, ret *Result) (*Result, error) {
	//@TODO
	// 1. 检查 ret.path 和 订阅列表
	// 2. 执行对应的订阅，
	// 3. dispatch.send()
	return nil, nil
}

func (e *Container) handleCallback(ctx context.Context, event *ContainerEvent, ret *Result) (*Result, error) {
	//@TODO 处理回调
	// 1. dispatch.respose(ret)
	fmt.Println("handleCallback", event.Callback)
	return nil, nil
}

func (e *Container) Entity(id string) (Entity, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if state, ok := e.entities[id]; ok {
		return state, nil
	}

	return nil, nil
}

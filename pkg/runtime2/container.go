/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runtime2

import (
	"context"
	"fmt"
	"sync"
)

//Container 作为 Runtime 负责 Entity 的生命周期管理
//Container 初始化时
//1. 初始化 Inbox(partition)，注册 HandleEvent
//

//2. HandleEvent 消息处理逻辑
//2.1 实体必须包含 entityID，创建、删除等消息：由 Container 处理
//    实体配置重载？Mapper变化了（Mapper包括 订阅-source、执行-target）
//
//升级Mapper执行的环境
//2.1 Cache 消息直接更新 Container 的 caches
//2.3 实体消息：首先查找 StateMachine，如果没有初始化,且加入map
//              entities[entityID]=StateMachine(entityID,dispatch,stateBytes)
//2.3 更新实体，记录下变更    -> StateResult
//
//3  HandleEvent 处理完毕后后
//2.2  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
//              mappers[entityID]=Mapper(dispatch,stateBytes)
//3.3 触发对应的Mapper（执行-target） -> Patch
//3.4 更新实体（target），记录下变更   -> StateResult(a.p2)
//
//4.收尾
//4.1 依照订阅发布实体变更  handlePatch（StateResult(a.p2)）
//4.2 处理 API 回调       handleCallback（StateResult(a.p2)）
//
//5. 初始化 StateMachine
//5.1 从 StateMachine 中读取 stateBytes
//5.2 从 stateBytes 中新建 StateMachine
//
//Q:
//Container 需要知道Event类型
//- 系统消息
//- 有没有回调
//- 是否为cache更新
// ContainerEvent
// - Manger\Entity\Cache
//
//如下处理流程
//insert into a select b.p1 as p2
//insert into c select a.p2 as p2
//
type Container struct {
	Index    int
	caches   map[string]Entity  //存放其他Container的实体
	entities map[string]Entity  //存放Container的实体
	mappers  map[string]*Mapper //存放Container的Mapper
	dispatch *Dispatch
	dao      Dao
	//kafka client
	//inbox    Inbox

	lock sync.RWMutex
}

type EventType string
type ContainerEventType string

const (
	OpContainer    EventType          = "core.operation.Container"
	OpEntity       EventType          = "core.operation.Entity"
	OpCache        EventType          = "core.operation.Cache"
	OpMapperCreate ContainerEventType = "core.operation.Mapper.Create"
	OpMapperUpdate ContainerEventType = "core.operation.Mapper.Update"
	OpMapperDelete ContainerEventType = "core.operation.Mapper.Delete"
	OpEntityCreate ContainerEventType = "core.operation.Entity.Create"
	OpEntityUpdate ContainerEventType = "core.operation.Entity.Update"
	OpEntityDelete ContainerEventType = "core.operation.Entity.Delete"
)

//处理消息
type ContainerEvent struct {
	ID       string
	Type     EventType
	Callback string
	Value    interface{}
}

type Entity interface {
	Handle(ctx context.Context, message interface{}) (*StateResult, error)
	Raw() ([]byte, error)
}

func NewContainer(partitionID int) *Container {
	return &Container{
		Index:    partitionID,
		caches:   map[string]Entity{},
		entities: map[string]Entity{},
		mappers:  map[string]*Mapper{},
		lock:     sync.RWMutex{},
	}
}

func (e *Container) DeliveredEvent(ctx context.Context, event interface{}) error {
	// 1. 通过 inbox 实现event 转换.
	panic("implement me.")
}

func (e *Container) HandleEvent(ctx context.Context, event *ContainerEvent) (*StateResult, error) {
	var (
		ret *StateResult
		err error
	)

	//1.BeforeProcess
	//1. 升级执行的环境
	//1.1 处理 Entity 的创建、删除
	//1.2 Cache 消息直接更新 Container 的 caches
	//1.3 更新实体，记录下变更    -> StateResult
	ret, err = e.UpdateWithEvent(ctx, event)
	if err != nil {
		return nil, err
	}

	//2.Process
	//2.1  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
	//              mappers[entityID]=Mapper(dispatch,stateBytes)
	//2.2 触发对应的Mapper（执行-target） -> Patch
	//2.3 更新实体（target），记录下变更   -> StateResult(a.p2)
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

func (e *Container) UpdateWithEvent(ctx context.Context, event *ContainerEvent) (*StateResult, error) {
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
func (e *Container) handleMangerEvent(ctx context.Context, event *ContainerEvent) (*StateResult, error) {
	panic("implement me")
}

//处理Cache
func (e *Container) handleCacheEvent(ctx context.Context, event *ContainerEvent) (*StateResult, error) {
	panic("implement me")
}

//Container 处理 event
func (e *Container) Process(ctx context.Context, event *ContainerEvent) (*StateResult, error) {
	panic("implement me")
	//2.2  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
	//              mappers[entityID]=Mapper(dispatch,stateBytes)
	//3.3 触发对应的Mapper（执行-target） -> Patch
	//3.4 更新实体（target），记录下变更   -> StateResult(a.p2)
	//EntityID := event.ID
	//mapper, err := e.Mapper(EntityID)
	//if err != nil {
	//	return nil, err
	//}
	//ret, err := mapper.Handle(ctx, event)
	//return nil, nil
}

//处理订阅
func (e *Container) handleSubscribe(ctx context.Context, ret *StateResult) (*StateResult, error) {
	//@TODO
	// 1. 检查 ret.path 和 订阅列表
	// 2. 执行对应的订阅，
	// 3. dispatch.send()
	return nil, nil
}

func (e *Container) handleCallback(ctx context.Context, event *ContainerEvent, ret *StateResult) (*StateResult, error) {
	//@TODO 处理回调
	// 1. dispatch.respose(ret)
	fmt.Println("handleCallback", event.Callback)
	return nil, nil
}

func (e *Container) Entity(id string) (Entity, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if entity, ok := e.entities[id]; ok {
		return entity, nil
	}
	//@TODO 初始化
	entity := NewEntity()
	e.entities[id] = entity
	return entity, nil
}

func (e *Container) Mapper(id string) (*Mapper, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if mapper, ok := e.mappers[id]; ok {
		return mapper, nil
	}
	//@TODO 初始化
	mapper := &Mapper{}
	e.mappers[id] = mapper
	return mapper, nil
}

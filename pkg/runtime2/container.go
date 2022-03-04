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

import "context"

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
//              machines[entityID]=StateMachine(entityID,dispatch,stateBytes)
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
	//inbox    Inbox
	caches   map[string]*TEntity //存放其他Container的实体
	machines map[string]*TEntity //存放Container的实体
	mappers  map[string]*Mapper  //存放Container的Mapper
	dispatch *Dispatch
	dao      Dao
	//kafka client
}

//处理消息
type ContainerEvent struct {
	Type  string
	Value interface{}
	//ID
	//TYPE = Manger\Entity\Cache
}

func (e *Container) DeliveredEvent(ctx context.Context, event interface{}) error {
	// 1. 通过 inbox 实现event 转换.
	panic("implement me.")
}

func (e *Container) HandleEvent(ctx context.Context, event ContainerEvent) (*StateResult, error) {
	var EntityID string
	//2.1 实体必须包含 entityID，创建、删除等消息：由 Container 处理
	//    实体配置重载？Mapper变化了（Mapper包括 订阅-source、执行-target）
	switch event.Type {
	case "Manger":
		//处理实体生命周期
	case "Entity":
		//处理实体状态
	case "Cache":
		//处理Cache
	}
	//升级Mapper执行的环境
	//2.1 Cache 消息直接更新 Container 的 caches
	//2.3 实体消息：首先查找 StateMachine，如果没有初始化,且加入map
	//              machines[entityID]=StateMachine(entityID,dispatch,stateBytes)
	ret, _ := e.machines[EntityID].Handle(ctx, event)
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

	e.handlePatch(ctx, ret)
	e.handleCallback(ctx, event, ret)
	return nil, nil
}

func (e *Container) ExecMapper(ctx context.Context, event *ContainerEvent) (*StateResult, error) {
	//2.2  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
	//              mappers[entityID]=Mapper(dispatch,stateBytes)
	//3.3 触发对应的Mapper（执行-target） -> Patch
	//3.4 更新实体（target），记录下变更   -> StateResult(a.p2)
	return nil, nil
}

func (e *Container) handlePatch(ctx context.Context, ret *StateResult) (*StateResult, error) {
	//@TODO
	// 1. 检查 ret.path 和 订阅列表
	// 2. 执行对应的订阅，
	// 3. dispatch.send()
	return nil, nil
}

func (e *Container) handleCallback(ctx context.Context, event ContainerEvent, ret *StateResult) (*StateResult, error) {
	//@TODO
	// 1. dispatch.respose(ret)
	return nil, nil
}

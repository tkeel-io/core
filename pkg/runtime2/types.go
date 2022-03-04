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

//Manager 负责当前服务的 Runtime
// CORE: partitionNum := 9
// Kafka: core_0\core_1....core_8
// Manger :3
// Manage1: core_0~core_2
// Manage2: core_3~core_5
// Manage3: core_6~core_8
//1. 订阅ETCD（Mapper）
//      MapperKey：TKELL-CORE/0/entityid/MapID
//      kafka://aaaaa.com/topic_<partitionID>
//1.1 partitionID := hash(entityid)%partitionNum
//2. 管理Container
//3. 消费Kafka，按照Kafka的 partition 来创建 Container(partition,dispatch,dao)
//
//Q:
//1.为什么不是 Container 消费Kafka，应为Kafka 的client无法控制只消费某个partition
//2.dao在哪个层次？
type Manager struct {
	containers map[string]Container
	dispatch   *Dispatch
	dao        Dao
}

//Container 作为 Runtime 负责 Entity 的生命周期管理
//Container 初始化时
//1. 初始化 Inbox(partition)，注册 MessageHandle
//
//2. MessageHandle 消息处理逻辑
//2.1 实体必须包含 entityID，创建、删除等消息：由 Container 处理
//    实体配置重载？Mapper变化了（Mapper包括 订阅-source、执行-target）
//
//升级Mapper执行的环境
//2.1 Cache 消息直接更新 Container 的 caches
//2.3 实体消息：首先查找 StateMachine，如果没有初始化,且加入map
//              machines[entityID]=StateMachine(entityID,dispatch,stateBytes)
//2.3 更新实体，记录下变更
//
//3  MessageHandle 处理完毕后后
//2.2  Mapper处理：首先查找 Mapper，如果没有初始化,且加入map
//              mappers[entityID]=Mapper(dispatch,stateBytes)
//3.3 触发对应的Mapper（执行-target）
//3.4 更新实体（target），记录下变更
//
//4.收尾
//4.1 依照订阅发布实体变更
//4.2 处理 API 回调
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
type Container struct {
	//inbox    Inbox
	caches   map[string]*StateMachine //存放其他Container的实体
	machines map[string]*StateMachine //存放Container的实体
	mappers  map[string]*Mapper       //存放Container的Mapper
	dispatch *Dispatch
	dao      Dao
	//kafka client
}

//Inbox 负责消费Kafka的单个 partition 消息，维护该partition的Ack与Offset
type Inbox struct {
	handle MessageHandle
}

//Mapper 负责处理对应的实体状态更新
type Mapper struct {
}

//StateMachine 作为 Entity 负责接收消息并处理自身状态
// 处理 EntityEvent
type StateMachine struct {
	handle MessageHandle
}

//Dispatch功能如下：
//1. 通过Dapr消费外部消息，转化为内部消息输出到对应的Kafka的partition
//2. 接受API的请求，转化为内部消息输出到对应的Kafka的partition
//3. 接收内部输出，并输出到对应的Kafka的partition
//4. 通过Dapr讲对外消息输出到 pubsub
//5. 接受API的回调，回调API网关
//提供 MessageHandle 接口
type Dispatch struct {
	//dapr client
	//kafka client
}

//Dao 负责读取实体状态，Dao不负责具体的序列化，提供 KV 存储。
type Dao struct {
}


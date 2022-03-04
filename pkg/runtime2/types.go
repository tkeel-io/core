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

	"github.com/tkeel-io/tdtl"
)

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

//Inbox 负责消费Kafka的单个 partition 消息，维护该partition的Ack与Offset
type Inbox struct {
	handle MessageHandle
}

//Mapper 负责处理对应的实体状态更新
type Mapper struct {
}

//StateMachine 作为 Entity 负责接收消息并处理自身状态
// 处理 EntityEvent
type EventHandle func(ctx context.Context, message interface{}) (*StateResult, error)
type Patch struct {
	JSONPath string
	OP       string
	Value    []byte
}

//Feed 包含实体最新状态以及变更
type StateResult struct {
	State  []byte
	Patchs []Patch
}

type TEntity struct {
	ID            string               `json:"id" msgpack:"id" mapstructure:"id"`
	Type          string               `json:"type" msgpack:"type" mapstructure:"type"`
	Owner         string               `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source        string               `json:"source" msgpack:"source" mapstructure:"source"`
	Version       int64                `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime      int64                `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	TemplateID    string               `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	Properties    map[string]tdtl.Node `json:"-" msgpack:"-" mapstructure:"-"`
	ConfigBytes   []byte               `json:"-" msgpack:"config_bytes" mapstructure:"config_bytes"`
	PropertyBytes []byte               `json:"property_bytes" msgpack:"property_bytes" mapstructure:"property_bytes"`
}

func (e *TEntity) Handle(ctx context.Context, message interface{}) (*StateResult, error) {
	//@TODO config
	//message -> Patch
	//e.update(Patch)
	return nil, nil
}

func (e *TEntity) RawByte(ctx context.Context) ([]byte, error) {
	return nil, nil
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

// 原有结构体
/*
type PatchData struct {
	Path     string      `json:"path"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type Entity struct {
	ID            string               `json:"id" msgpack:"id" mapstructure:"id"`
	Type          string               `json:"type" msgpack:"type" mapstructure:"type"`
	Owner         string               `json:"owner" msgpack:"owner" mapstructure:"owner"`
	Source        string               `json:"source" msgpack:"source" mapstructure:"source"`
	Version       int64                `json:"version" msgpack:"version" mapstructure:"version"`
	LastTime      int64                `json:"last_time" msgpack:"last_time" mapstructure:"last_time"`
	TemplateID    string               `json:"template_id" msgpack:"template_id" mapstructure:"template_id"`
	ConfigBytes   []byte               `json:"-" msgpack:"config_bytes" mapstructure:"config_bytes"`
	PropertyBytes []byte               `json:"property_bytes" msgpack:"property_bytes" mapstructure:"property_bytes"`
}

type ItemsData struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Owner        string   `json:"owner"`
	Source       string   `json:"source"`
	PropertyKeys []string `json:"property_keys"`
}

{
OP: Props/Config - Update\Replace\GET
Path:
Value:
}

*/

## Core Documentation



<h1 align="center"> tKeel-Core</h1>
<h5 align="center"> The digital engine of world</h5>
<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/tkeel-io/core)](https://goreportcard.com/report/github.com/tkeel-io/core)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tkeel-io/core)
![GitHub](https://img.shields.io/github/license/tkeel-io/core?style=plastic)
[![GoDoc](https://godoc.org/github.com/tkeel-io/core?status.png)](http://godoc.org/github.com/tkeel-io/core)
</div>

🌰 Core 是 tKeel 物联网平台的数据中心，高性能、可拓展的轻量级下一代数字化数据引擎。

以实体（entity）为操作单元，通过简易明了的 API 对外提供读写能力（属性读写、时序查询、订阅，映射等）。

[English](README.md)

## 🪜 架构设计
架构按操作分为分为了两个平面。

- **控制**： 通过 core 向外暴露的 APIs 向 core 发送控制请求（如实体，映射，订阅的创建等）。
- **数据**： 在两个通信服务节点之间建立直连的 [channel](docs/channel/channel.md)，避免由网关和边车带来的长链路路由延迟，实现高性能的数据交互。


<div align="center">

![img.png](images/architecture.png)
<i>架构图 </i>
</div>



## 🌱 基本概念
### 实体（Entity）
实体是我们在物联网世界中对 Things 的一种抽象，是 Core 操作的基础对象。包括智能灯、空调、网关，房间，楼层，甚至是通过数据聚合生成的虚拟设备等等，我们将这些 `Things` 进行抽象，
定义为实体。

*属性* 是对实体某种信息的描述，一个实体包含两类属性
1. **基础属性**: 每个实体都必备的属性，如 `id`，`owner`等用于标识实体共有特征的属性。
2. **扩展属性**: 实体除基础属性外的属性，这种属性属于某一类或某一个实体的特征描述，比如一个 **温度计** 的温度。


更多设计细节请阅读[实体文档](docs/entity/entity.md)

### Actor
[Actor](docs/actors/actor.md) 是实体（Entity）的运行时的一种模式抽象, 用于维护实体的实时状态以及提供实体的一些具体行为。

### 映射
[映射](docs/mapper/mapper.md) 是实体之间数据传递和映射的一种规则的定义，用于实现上报数据的向上传播以及控制命令的向下传播。  
<div align="center">

![img.png](images/message_passing.png)

<i>映射模拟</i>
</div>

上图中蓝色线条代表数据的上行，如设备数据上报，黑色代表数据的下行，如指令数据的下行。



映射操作的执行包含两步:

1. 写复制: 实现实体属性变更时，将变更向下游实体传递。
2. 计算更新: 对上游实体产生的变更组合计算，然后将计算结果更新到当前实体。


<div align="center">

![img.png](images/mapping.png)
</div>


### 关系

在物理世界中，实体与实体之间往往不是相互孤立的，它们之间往往存在各式各样的联系，如交换机，路由器，终端设备，服务器通过光纤连接，在网络拓扑图中这些设备实体有`连接关系`。这些关系将这些独立的设备实体链接在一起，组成复杂而精密的网络，向外提供稳定而高速的网络通信服务。当然实体不局限于设备实体，关系也不仅仅局限于 `连接关系`，[更多设计细节请阅读关系文档](docs/relationship/relationship.md)。



### 模型


我们将实体属性的约束集合定义为模型。实体是属性数据的载体，但是如何解析和使用实体的属性数据，我们需要实体属性的描述信息，如类型，取值范围等，我们将这些描述信息称之为 `约束`。而模型就是一个包含`约束`集合的载体，模型也以实体的形式存在， [更多设计细节请阅读模型文档](docs/model/model.md) 。



### 订阅
Core 提供了简捷方便的[订阅](docs/subscription/subscription.md) ，供开发者实时获取自己关心的数据。

在 tKeel 平台中用于多个 plugin 之间和一个 plugin 内所有以实体为操作对象的数据交换。

底层实现逻辑是这样的：每个 plugin 在注册的时候在 Core 内部自动创建一个交互的 `pubsub`，名称统一为 pluginID-pubsub,
订阅的 `topic` 统一为 pub-core，sub-core，只有 core 与该 plugin 有相关权限
比如
iothub: iothub-pubsub

**订阅** 分为三种：
- **实时订阅**： 订阅将实体的实时数据发送给订阅者。
- **变更订阅**： 订阅者订阅的实体属性发生变更且满足变更条件时，订阅将实体属性数据发送给订阅者。
- **周期订阅**： 订阅周期性的将实体属性数据发送给订阅者。






## 目录

- **[Intropduction](introduction/quick-start.md)**
- **[Development](development/README.md)**
- **[APIs](api/index.md)**
- **[Entity](entity/entity.md)**
- **[Entity Runtime](actors/actor.md)**
    - **[Time Series Store](actors/time-series-store.md)**
- **[Mapper](mapper/mapper.md)**
- **[TQL](tql/tql.md)**
- **[Model](model/model.md)**
- **[Subscription](subscription/subscription.md)**
- **[Relationship](relationship/relationship.md)**
- **[Channel](channel/channel.md)**
- **[Inbox](inbox/inbox.md)**
- **[Distributed](distribute/distributed.md)**







## ☎️ 联系我们
提出您可能有的任何问题，我们将确保尽快答复！

| 平台 | 链接 |
|:---|----|
|email| tkeel@yunify.com|
|微博| [@tkeel]()|


## 🏘️ 仓库

| 仓库 | 描述 |
|:-----|:------------|
| [tKeel](https://github.com/tkeel-io/tkeel) | tKeel 开放物联网平台|
| [Core](https://github.com/tkeel-io/core) | tKeel 的数据中心 |
| [CLI](https://github.com/tkeel-io/cli) | tKeel CLI 是用于各种 tKeel 相关任务的主要工具 |
| [Helm](https://github.com/tkeel-io/helm-charts) | tKeel 对应的 Helm charts |


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

## 🪜 架构设计
架构按操作分为分为了两个平面。

- **控制**： 通过 core 向外暴露的 APIs 向 core 发送控制请求（如实体，映射，订阅的创建等）。
- **数据**： 在两个通信服务节点之间建立直连的 [channel](docs/channel/channel.md)，避免由网关和边车带来的长链路路由延迟，实现高性能的数据交互。


<div align="center">

![img.png](docs/images/architecture.png)
<i>架构图 </i>
</div>


    
## 🚪 快速入门
Core 是 tKeel 的一个重要基础组件，拥有单独部署能力，使用相关特性做满足广大用户需求的功能也是我们竭力想要的。

### 安装需要
🔧 在使用 Core 之前请先确保你做足了准备。 
1. [Kubernetes](https://kubernetes.io/)
2. [Dapr with k8s](https://docs.dapr.io/getting-started/)


### 通过 tKeel 安装
Core 作为 tKeel 的基础组件，相关 API 的调用均通过 tKeel 代理实现。（详细请见[tKeel CLI 安装文档](https://github.com/tkeel-io/cli )）

### 独立部署
拉取仓库
```bash 
$ git clone  git@github.com:tkeel-io/core.git
$ cd core
```
#### Self-hosted
> ⚠️ 注意：请本地先运行一个 redis 进程，监听 6379 端口，无密码
##### 通过 Dapr 启动项目
```bash
$ dapr run --app-id core --app-protocol http --app-port 6789 --dapr-http-port 3500 --dapr-grpc-port 50001 --log-level debug  --components-path ./examples/configs/core  go run . serve
```
#### Kubernetes
1. 部署 reids 服务
    ```bash
    $ helm install redis bitnami/redis
    ```
2. 运行 core 程序
    ```bash
    $ kubectl apply -f k8s/core.yaml
    ```

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

![img.png](docs/images/message_passing.png)

<i>映射模拟</i>
</div>

上图中蓝色线条代表数据的上行，如设备数据上报，黑色代表数据的下行，如指令数据的下行。



映射操作的执行包含两步:

1. 写复制: 实现实体属性变更时，将变更向下游实体传递。
2. 计算更新: 对上游实体产生的变更组合计算，然后将计算结果更新到当前实体。


<div align="center">

![img.png](docs/images/mapping.png)
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


### 作为 tKeel 组件运行
#### 示例
在 tKeel 相关组件安装完成之后，[Python 示例](examples/iot-paas.py) 展示了生成 MQTT 使用的 `token`，然后创建实体，上报属性，获取快照，订阅实体的属性等功能。  
为了方便说明，下面是我们使用外部流量方式访问 Keel，和 Python 作为示例语言的代码。我们需要keel和mqtt broker的服务端口用于演示。

##### 获取服务端口
1. Keel 服务端口
```bash
$ KEEL_PORT=$(kubectl get -o jsonpath="{.spec.ports[0].nodePort}" services keel)
```
2. MQTT Server 服务端口
```bash
$ MQTT_PORT=$(kubectl get -o jsonpath="{.spec.ports[0].nodePort}" services emqx)
```

keel openapi 服务地址为k8s ip:keel暴露的nodeport端口
```python
// Source: examples/iot-paas.py
keel_url = "http://{host}:{port}/v0.1.0"
```

##### 创建 token
```python
// Source: examples/iot-paas.py
def create_entity_token(entity_id, entity_type, user_id):
    data = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id)
    token_create = "/auth/token/create"
    res = requests.post(keel_url + token_create, json=data)
    return res.json()["data"]["entity_token"]
```

##### 创建实体
```python
// Source: examples/iot-paas.py
def create_entity(entity_id, entity_type, user_id, plugin_id, token):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, source="abc", plugin_id=plugin_id)
    entity_create = "/core/plugins/{plugin_id}/entities?id={entity_id}&type={entity_type}&owner={user_id}&source={source}".format(
        **query)
    data = dict(token=token)
    res = requests.post(keel_url + entity_create, json=data)
    print(res.json())
```

##### 上报实体属性
```python
// Source: examples/iot-paas.py
def on_connect(client, userdata, flags, rc):
    if rc == 0:
        print("Connected to MQTT Broker!")
    else:
        print("Failed to connect, return code %d\n", rc)

client = mqtt_client.Client(entity_id)
client.username_pw_set(username=user_id, password=token)
client.on_connect = on_connect
client.connect(host=broker, port=port)
client.loop_start()
time.sleep(1)
payload = json.dumps(dict(p1=dict(value=random.randint(1, 100), time=int(time.time()))))
client.publish("system/test", payload=payload)
```

##### 获取实体快照
```python
// Source: examples/iot-paas.py
def get_entity(entity_id, entity_type, user_id, plugin_id):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, plugin_id=plugin_id)
    entity_create = "/core/plugins/{plugin_id}/entities/{entity_id}?type={entity_type}&owner={user_id}&source={plugin_id}".format(
        **query)
    res = requests.get(keel_url + entity_create)
    print(res.json()["properties"])

```

##### 订阅实体
运行订阅实体之前，先要创建订阅目的地的pubsub，可以通过运行消费topic的示例yaml创建[消费示例yaml](examples/subclient/client.yaml)

```python
// Source: examples/iot-paas.py
def create_subscription(entity_id, entity_type, user_id, plugin_id, subscription_id):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, source="abc", plugin_id=plugin_id, subscription_id=subscription_id)
    entity_create = "/core/plugins/{plugin_id}/subscriptions?id={subscription_id}&type={entity_type}&owner={user_id}&source={source}".format(
        **query)
    data = dict(mode="realtime", source="ignore", filter="insert into abc select " + entity_id + ".p1", target="ignore", topic="abc", pubsub_name="client-pubsub")
    print(data)
    res = requests.post(keel_url + entity_create, json=data)
    print(res.json())
```

##### 消费 topic 数据
消费程序作为一个独立的app消费相关topic数据并展示[消费示例](examples/subclient)
```python
// Source: examples/subclient/app.py
import flask
from flask import request, jsonify
from flask_cors import CORS
import json
import sys

app = flask.Flask(__name__)
CORS(app)

@app.route('/dapr/subscribe', methods=['GET'])
def subscribe():
    subscriptions = [{'pubsubname': 'client-pubsub',
                      'topic': 'abc',
                      'route': 'data'}]
    return jsonify(subscriptions)

@app.route('/data', methods=['POST'])
def ds_subscriber():
    print(request.json, flush=True)
    return json.dumps({'success':True}), 200, {'ContentType':'application/json'}
app.run()
```

### Entity 示例
因为当前 Dapr SDK 不能处理 HTTP 请求中的 Header，参数通过 path 和 query 的方式传递。

[examples](examples/entity) 该示例中的功能，创建实体，通过 pubsub 更新实体属性，查询实体。

#### 创建实体
```go
    // Source: examples/entity/main.go
	client, err := dapr.NewClient()
	if nil != err {
		panic(err)
	}

	// create entity.
	createUrl := "plugins/pluginA/entities?id=test1&owner=abc&source=abc&type=device"

	result, err := client.InvokeMethodWithContent(context.Background(),
		"core",
		createUrl,
		"POST",
		&dapr.DataContent{
			ContentType: "application/json",
		})
	if nil != err {
		panic(err)
	}
	fmt.Println(string(result))
```
#### 更新实体属性
```go
    // Source: examples/entity/main.go
    data := make(map[string]interface{})
	data["entity_id"] = "test1"
	data["owner"] = "abc"
	dataItem := make(map[string]interface{})
	dataItem["core"] = ValueType{Value: 189, Time: time.Now().UnixNano() / 1e6}
	data["data"] = dataItem

	err = client.PublishEvent(context.Background(),
		"client-pubsub",
		"core-pub",
		data,
	)

	if nil != err {
		panic(err)
	}
```

#### 获取实体属性
```go
    // Source: examples/entity/main.go
    getUrl := "plugins/pluginA/entities/test1?owner=abc&source=abc&type=device"

	result, err = client.InvokeMethodWithContent(context.Background(),
		"core",
		getUrl,
		"GET",
		&dapr.DataContent{
			ContentType: "application/json",
		})
	if nil != err {
		panic(err)
	}
	fmt.Println(string(result))
```


## ⚙️ API
Core 的更多功能 API 详细请参见[ API 文档](docs/api/index.md)

## 💬 一起点亮世界
如果您有任何的建议和想法，欢迎您随时开启一个 [Issue](https://github.com/tkeel-io/core/issues )，期待我们可以一起交流，让世界更美好。

同时 **非常感谢** 您的 `反馈` 与 `建议` ！

[社区文档](docs/development/README.md) 将会带领您了解如何开始为 tKeel 贡献。

### 🧱 贡献一己之力

[开发指南](docs/development/developing-tkeel.md) 向您解释了如何配置您的开发环境。

我们有这样一份希望项目参与者遵守的 [行为准则](docs/community/code-of-conduct.md)。请阅读全文，以便您了解哪些行为会被容忍，哪些行为不会被容忍。

### ☎️ 联系我们
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


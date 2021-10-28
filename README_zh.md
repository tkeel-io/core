<h1 align="left"> tKeel-Core </h1>
🌰 tKeel 物联网平台的数据中心。将世界万物数字化的数据库。

[comment]: <> (🌰 tKeel 物联网平台的数据中心。将世界万物抽象成类似于元宇宙的一个数值化高拓展性的数据库。)

Core 以实体（entity）为操作对象，通过简易明了的 API 对外提供读写能力（属性读写、时序查询、订阅等）。

[English](README.md)

## 架构设计
架构按操作分为 **控制平面** 和 **数据平面**。

- 控制平面：
  <br> 通过 http 进行实体的创建查询等操作，
- 数据平面：
  <br> 通过 dapr 的 pubsub 完成数据的高效读写与订阅。

架构图：
![img.png](docs/images/architecture.png)

    
## 快速入门
Core 是 tKeel 的一个重要基础组件，拥有单独部署能力，使用相关特性做满足广大用户需求的功能也是我们竭力想要的。

### 安装需要
1. [Kubernetes](https://kubernetes.io/)
2. [Dapr](https://docs.dapr.io/getting-started/)


### 通过 tKeel 安装
Core 作为 tKeel 的基础组件，相关 API 的调用均通过 tKeel 代理实现。（详细请见[tKeel CLI 安装文档](https://github.com/tkeel-io/cli )）

### 独立部署
通过 Dapr 启动该项目。

1. 拉去仓库
```bash 
$ git clone  git@github.com:tkeel-io/core.git
```
2. 启动程序
```bash
 
```

## 基本概念
### 实体（Entity）
实体是我们在物联网世界中对 Things 的一种抽象，是所有操作的基础对象。包括网关、设备、关于设备的聚合等概念，都进行了抽象，
抽象出来了这样一个实体的概念。

*属性* 是对实体某种信息的描述，一个实体包含三类属性
1. **基础属性** 每个实体都必备的属性，如 `owner`，`plugin`
2. **自身属性** 实体自身的属性，比如一个 **温度计** 的 `温度`
3. **映射属性** 由其他实体属性映射而来的一个缓存聚合属性。

更多设计细节请阅读[实体文档](docs/entity/entity.md)

### Actor
[Actor](docs/actors/actor.md) 是 实体（Entity）的运行时的一种模式抽象, 用于维护实体的实时状态以及提供实体的一些具体行为。

### 关系
关系是实体与实体之间的联系。


### 映射
映射是实体属性的传播，可以实现上报数据的向上传播以及控制命令的向下传播。  
![img.png](docs/images/message_passing.png)

蓝色线条代表上行，黑色代表下行

映射的操作包含两个部分: 写复制和计算更新  
![img.png](docs/images/mapping.png)

参见[映射](docs/mapper/mapper.md)
### 模型
模型是用来约束实体属性的定义。
有模型的实体属性需要按照模型的要求对值进行处理，比如需要进时序数据库时或者需要用于搜索等。

### 订阅
Core 提供了简捷方便的订阅功能，供开发者实时获取自己关心的数据。

在 tKeel 平台中用于多个 plugin 之间和一个 plugin 内所有以实体为操作对象的数据交换。

底层实现逻辑是这样的：每个 plugin 在注册的时候在 Core 内部自动创建一个交互的 `pubsub`，名称统一为 pluginID-pubsub,
订阅的 `topic` 统一为 pub-core，sub-core，只有 core 与该 plugin 有相关权限
比如
iothub: iothub-pubsub

**订阅** 分为三种：
- **实时订阅**： 收到消息时触发 
- **变更订阅**： 实体属性有变动时触发 
- **周期订阅**： 周期性触发

详细请参见[订阅文档](docs/subscription/subscription.md)

### 作为 tKeel 组件运行
#### 示例
在 tKeel 相关组件安装完成之后，[Python 示例](examples/iot-paas.py) 展示了生成 MQTT 使用的 `token`，然后创建实体，上报属性，获取快照，订阅实体的属性等功能。  
为了方便说明，下面是我们使用外部流量方式访问 Keel，和 Python 作为示例语言的代码。

##### 获取服务端口
1. Keel 服务端口
```bash
KEEL_PORT=$(kubectl get -o jsonpath="{.spec.ports[0].nodePort}" services keel)
```
2. MQTT Server 服务端口
```bash
MQTT_PORT=$(kubectl get -o jsonpath="{.spec.ports[0].nodePort}" services emqx)
```

keel openapi 服务地址为k8s ip:keel暴露的nodeport端口
```python
// examples/iot-paas.py
keel_url = "http://{host}:{port}/v0.1.0"
```

##### 创建 token
```python
// examples/iot-paas.py
def create_entity_token(entity_id, entity_type, user_id):
    data = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id)
    token_create = "/auth/token/create"
    res = requests.post(keel_url + token_create, json=data)
    return res.json()["data"]["entity_token"]
```

##### 创建实体
```python
// examples/iot-paas.py
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
// examples/iot-paas.py
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
// examples/iot-paas.py
def get_entity(entity_id, entity_type, user_id, plugin_id):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, plugin_id=plugin_id)
    entity_create = "/core/plugins/{plugin_id}/entities/{entity_id}?type={entity_type}&owner={user_id}&source={plugin_id}".format(
        **query)
    res = requests.get(keel_url + entity_create)
    print(res.json()["properties"])

```

##### 订阅实体
##### 消费topic数据

### Entity 示例
因为当前 Dapr SDK 不能处理 HTTP 请求中的 Header，参数通过 path 和 query 的方式传递。

[examples](examples/entity) 该示例中的功能，创建实体，通过 pubsub 更新实体属性，查询实体。

#### 创建实体
```go
    // examples/entity/main.go
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
    // examples/entity/main.go
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
    // examples/entity/main.go
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


## API
Core 的更多功能 API 详细请参见[ API 文档](docs/api/index.md)

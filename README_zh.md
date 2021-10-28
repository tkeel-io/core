# Core
[English](README.md)

Core通过api以实体为对象对外提供属性搜索，时序查询，数据写入，数据查询，数据订阅等服务，按操作分为控制平面和数据平面。  
控制平面通过http进行实体的创建查询等操作，数据平面通过dapr的pubsub完成数据的高效写入和订阅。

![img.png](docs/images/architecture.png)

    
## 快速入门
core可以作为tkeel的一个基础组件运行，也可以部署为一个单独的服务。

### 作为tkeel的组件运行

core作为tkeel的基础组件，相关API的调用需要通过tkeel代理。参见tkeel文档。    
在tkeel相关组件完成之后，我们可以生成用于mqtt使用的token，创建实体，上报属性，获取快照，订阅实体的属性等。  
为了方便说明，我们使用外部流量方式访问keel，使用python作为示例代码语言。
[code](examples/iot-paas.py)
#### 获取服务端口
1. keel服务端口
```bash
KEEL_PORT=$(kubectl get -o jsonpath="{.spec.ports[0].nodePort}" services keel)
```
2. mqtt server服务端口
```bash
MQTT_PORT=$(kubectl get -o jsonpath="{.spec.ports[0].nodePort}" services emqx)
```

keel openapi 服务地址为k8s ip:keel暴露的nodeport端口
```python
keel_url = "http://{host}:{port}/v0.1.0"
```

#### 创建token

```python
def create_entity_token(entity_id, entity_type, user_id):
    data = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id)
    token_create = "/auth/token/create"
    res = requests.post(keel_url + token_create, json=data)
    return res.json()["data"]["entity_token"]
```

#### 创建实体
```python
def create_entity(entity_id, entity_type, user_id, plugin_id, token):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, source="abc", plugin_id=plugin_id)
    entity_create = "/core/plugins/{plugin_id}/entities?id={entity_id}&type={entity_type}&owner={user_id}&source={source}".format(
        **query)
    data = dict(token=token)
    res = requests.post(keel_url + entity_create, json=data)
    print(res.json())
```

#### 上报实体属性
```python
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

#### 获取实体快照
```python
def get_entity(entity_id, entity_type, user_id, plugin_id):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, plugin_id=plugin_id)
    entity_create = "/core/plugins/{plugin_id}/entities/{entity_id}?type={entity_type}&owner={user_id}&source={plugin_id}".format(
        **query)
    res = requests.get(keel_url + entity_create)
    print(res.json()["properties"])

```

#### 订阅实体
#### 消费topic数据


### 独立部署
当前dapr sdk不能处理http请求中的header，参数通过path和query进行传递
示例程序的功能，创建实体，通过pubsub更新实体属性，查询实体。  
参见[examples](examples/entity/README.md)


## 基本概念
### 实体
实体作为物联网世界里的操作对象，包括网关，设备，设备的聚合抽象以及抽象出来的概念比如订阅等等。  
属性对一个实体某种信息的描述，一个实体包含三类属性  
1. 基础属性，每个实体都必备的属性，如owner，plugin
2. 自身属性，实体自己的属性，如一个温度计的温度
3. 映射属性，因为映射，订阅等关联操作，由其他实体写复制映射过来的属性

参见[实体](docs/entity/entity.md)

### actor
actor是 Entity的运行时模式, 用于维护Entity的实时状态和提供Entity的具体行为。

参见[actor](docs/actors/actor.md)

### 关系
关系是实体与实体之间的关系

### 映射
映射用来实现实体属性的传播，可以实现上报数据的向上传播以及控制命令的向下传播。  
![img.png](docs/images/message_passing.png)
 
 蓝色线条代表上行，黑色代表下行
 
映射的操作包含两个部分: 写复制和计算更新
![img.png](docs/images/mapping.png)

参见[映射](docs/mapper/mapper.md)
### 模型
模型用来约束实体的属性
有模型的属性需要按照模型的要求对属性的值进行处理，比如要进时序DB或者要用于搜索。

### 订阅
core提供的订阅用于plugin之间以及plugin内部以实体为对象的数据交换。  
每个plugin在注册的时候自动创建一个与core交互的pubsub，名称统一为pluginID-pubsub, topic统一为pub-core，sub-core，只有core与该plugin有相关权限
比如
iothub: iothub-pubsub

订阅分为三种：
1. 实时订阅（收到消息就触发）
2. 变更订阅（属性有变更时触发）
3. 周期订阅（周期性上报所有属性）

参见[订阅](docs/subscription/subscription.md)

## API

参见[API](docs/api/index.md)

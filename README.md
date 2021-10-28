<h1 align="left"> tKeel-Core </h1>
🌰 tKeel 物联网平台的数据中心。对世界万物的抽象数据库。

[中文文档](README_zh.md)


## Run

### Self-hosted
本地运行一个redis，监听6379端口，无密码  
```bash
dapr run --app-id core --app-protocol http --app-port 6789 --dapr-http-port 3500 --dapr-grpc-port 50001 --log-level debug  --components-path ./config go run . serve
```
执行测试

### Kubernetes
1. 部署reids服务

    ```bash
    helm install redis bitnami/redis
    ```
2. 部署pubsub和state组件 
    ```bash
    kubectl apply -f redis-state-core.yaml
    kubectl apply -f redis-pubsub-core.yaml
    kubectl apply -f binding-core.yaml
    ```
3. 运行core程序
    ```bash
    kubectl apply -f core.yaml
    ```
4. 测试  
    ```bash
    kubectl apply -f client.yaml

    kubectl get pod |grep client  // 找到对应的pod

    kubectl exec -it client-***-* -- /bin/sh // 对应的pod名称

    ```
    执行测试



### Required

1. 需要创建数据库`kcore`
```sql
CREATE DATABASE IF NOT EXISTS kcore 
	DEFAULT CHARACTER SET utf8;
```
1. 需要创建表`kcore.entity`
```sql
  CREATE TABLE IF NOT EXISTS entity(
    id varchar(127) UNIQUE NOT NULL,
    owner VARCHAR(63) NOT NULL,
    source VARCHAR(63) NOT NULL,
    tag VARCHAR(63),
    status VARCHAR(63),
    version INTEGER,
    entity_key VARCHAR(127),
    deleted_id VARCHAR(255),
    PRIMARY KEY ( id )
  )ENGINE=InnoDB DEFAULT CHARSET=utf8;
```


## Test

> api调用中，必须设置Header: Source, Owner, 可选字段： Type字段:
1. `Source`标识请求的发起者，如设备管理`device-management`，`Owner`标识是由哪一个用户发起的请求，`Type`标识实体类型。
2. 可以在http request的Header中设置`"Source":"abcd"`和`"Owner":"admin"`和`"Type":"DEVICE"`。
3. 可以在http request的Query中设置`source=abcd&owner=admin&type=DEVICE`。
4. 或者混合使用，Header中的设置覆盖Query中的设置。
5. `Type`字段在实体创建时是必选的。



### 创建 Entity

**Params：**
| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属Plugin。 | 
| EntityId | string | false | path/query | 用于标识创建的实体的Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|


创建entity, POST支持`upsert`操作
```bash
# 指定entityId创建entity
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123?owner=admin&type=DEVICE" \
  -H "Content-Type: application/json" \
  -H "Source: abcd" \
  -d '{
       "status": "completed"
     }'

# upsert
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123?source=abcd&owner=admin&type=DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
       "status": "start",
       "temp": 234
     }'

# 不指定entity id
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities" \
  -H "Source: abcd" \
  -H "Owner: admin" \
  -H "Type: DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
       "status": "completed"
     }'
```


### 查询 Entity

**Params：**
| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属Plugin。 | 
| EntityId | string | true | path/query | 实体的Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|

```bash
curl -X GET "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123" \
  -H "Source: abcd" \
  -H "Owner: admin"  \
  -H "Type: DEVICE"
```


### 更新 Entity

**Params：**
| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属Plugin。 | 
| EntityId | string | true | path/query | 实体的Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|

```bash
curl -X PUT "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123" \
  -H "Source: abcd" \
  -H "Owner: admin" \
  -H "Type: DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
       "status": "testing",
       "temp":123
     }'
```



### 删除 Entity


**Params：**
| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属Plugin。 | 
| EntityId | string | true | path/query | 实体的Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|

```bash
curl -X DELETE "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123" \
  -H "Source: abcd" \
  -H "Owner: admin"  \
  -H "Type: DEVICE" 
```


## Mapper

### Append 

curl -X PUT "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123/mappers" \
  -H "Source: abcd" \
  -H "Owner: admin" \
  -H "Type: DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
       "name": "subscribe-test234",
       "tql": "insert into test123 select test234.temp as temp"
     }'


## openapi 

```bash
# call /v1/identify
curl -X GET http://localhost:3500/v1.0/invoke/core/method/v1/identify

# call /v1/state
curl -X GET http://localhost:3500/v1.0/invoke/core/method/v1/status
```

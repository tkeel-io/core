# core

> core is the database for digital twins.

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
CREATE DATABASE IF NOT EXISTS test_db_char 
	DEFAULT CHARACTER SET utf8;
```
1. 需要创建表`kcore.entity`
```sql
  CREATE TABLE IF NOT EXISTS entity(
    id varchar(127) UNIQUE NOT NULL,
    user_id VARCHAR(63) NOT NULL,
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

> api调用中，必须设置Header: Source, User, Type字段:
1. 可以在http request的Header中设置`"Source":"abcd"`和`"User":"admin"`和`"Type":"DEVICE"`。
2. 可以在http request的Query中设置`source=abcd&user_id=admin&type=DEVICE`
3. 或者混合使用，Header中的设置覆盖Query中的设置。


### Entity

创建entity, POST支持`upsert`操作
```bash
# 指定entityId创建entity
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/entities?id=test123&tag=test&source=abcd&user_id=admin&type=device" \
  -H "Content-Type: application/json" \
  -d '{
       "status": "completed"
     }'

# upsert
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/entities?id=test123&tag=test&source=abcd&user_id=admin&type=device" \
  -H "Content-Type: application/json" \
  -d '{
       "status": "start",
       "temp": 234
     }'

# 不指定entity id
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/entities?tag=test" \
  -H "Source: abcd" \
  -H "User: admin" \
  -H "Type: DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
       "status": "completed"
     }'
```


### 查询entity
```bash
curl -X GET "http://localhost:3500/v1.0/invoke/core/method/entities?id=test123" \
  -H "Source: abcd" \
  -H "User: admin"  \
  -H "Type: DEVICE"
```


### 更新entity
```bash
curl -X PUT "http://localhost:3500/v1.0/invoke/core/method/entities?id=test123&tag=tomas" \
  -H "Source: abcd" \
  -H "User: admin" \
  -H "Type: DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
       "status": "testing",
       "temp":123
     }'
```



### 删除entity
```bash
curl -X DELETE "http://localhost:3500/v1.0/invoke/core/method/entities?id=test123&tag=tomas" \
  -H "Source: abcd" \
  -H "User: admin"  \
  -H "Type: DEVICE" 
```


## openapi 

```bash
# call /v1/identify
curl -X GET http://localhost:3500/v1.0/invoke/core/method/v1/identify

# call /v1/state
curl -X GET http://localhost:3500/v1.0/invoke/core/method/v1/status
```

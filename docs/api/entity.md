## Entity APIs



----
## Tips

> api调用中，必须设置Header: Source, Owner, 可选字段： Type字段:
1. `Source`标识请求的发起者，如设备管理`device-management`，`Owner`标识是由哪一个用户发起的请求，`Type`标识实体类型。
2. 可以在http request的Header中设置`"Source":"abcd"`和`"Owner":"admin"`和`"Type":"DEVICE"`。
3. 可以在http request的Query中设置`source=abcd&owner=admin&type=DEVICE`。
4. 或者混合使用，Header中的设置覆盖Query中的设置。
5. `Type`字段在实体创建时是必选的。



### 创建 Entity

- Method: **POST**
- URL: ```http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities?id={entity_id}&owner={owner}&type={type}```

**Params：**

| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属Plugin。 | 
| EntityId | string | false | path/query | 用于标识创建的实体的Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|
| Body |json|false|body|用于创建实体时的初始属性。|


创建entity, POST支持`upsert`操作
```bash
# 创建entity
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities?owner=admin&type=DEVICE" \
  -H "Content-Type: application/json" \
  -H "Source: abcd" \
  -d '{
       "status": "completed"
     }'

# 指定entityId创建entity
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities?id=test123&source=abcd&owner=admin&type=device" \
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

- Method: **GET**
- URL: ```http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities/{entity_id}?owner={owner}&type={type}```
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
- Method: **PUT**
- URL: ```http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities/{entity_id}?owner={owner}&type={type}```

**Params：**

| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属Plugin。 | 
| EntityId | string | true | path/query | 实体的Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|
| Body |json|false|body|用于更新的实体的属性|

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
- Method: **DELETE**
- URL: ```http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities/{entity_id}?owner={owner}&type={type}```

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
  -H "Owner: admin" \
  -H "Type: DEVICE" 
```



### 筛选 Entities
- Method: **GET**
- URL: ```http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities?owner={owner}&type={type}```

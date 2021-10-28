## Entity APIs

该文档整理所有 Entity 相关的 API。

----

## Tips

> 🚨 API 调用中，请先按规范做必要的相应设置

请求中必须有 `Source` 和 `Owner`； `Type` 字段在 Entity 创建时是必须项。

### 说明

`Source` 标识请求的发起者，如设备管理`device-management`，`Owner`标识是由哪一个用户发起的请求，`Type`标识实体类型。

### 示例

#### 通过请求 Header

在 HTTP 请求的 Header 中如下设置

- `"Source":"abcd"`
- `"Owner":"admin"`
- `"Type":"DEVICE"`

#### 通过请求 Query

在 HTTP 请求的 Query 中设置 `source=abcd&owner=admin&type=DEVICE`

#### 混合使用

可以混合以上两者使用，<u> 这样做的话 Header 中的配置信息会覆盖 Query 中的数据。</u>

### 创建 Entity

- Method: **POST**
- URL:

 ```
 http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities?id={entity_id}&owner={owner}&type={type}
```

**Params：**

| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属 Plugin。 | 
| EntityId | string | false | path/query | 用于标识创建的实体的Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起 Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|
| Body |json|false|body|用于创建实体时的初始属性。|

创建 entity, POST 支持 `upsert` 操作

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
- URL:

```
http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities/{entity_id}?owner={owner}&type={type}
```

**Params：**

| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属Plugin。 |
| EntityId | string | true | path/query | 实体的Id。|
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
- URL:

```
http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities/{entity_id}?owner={owner}&type={type}
```

**Params：**

| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属 Plugin。 | 
| EntityId | string | true | path/query | 实体的 Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起 Plugin。|
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
- URL:

```
http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities/{entity_id}?owner={owner}&type={type}
```

**Params：**

| Name | Type | Required | Where | Description |
| ---- | ---- | -------- | ----- | ----------- |
| PluginId | string | true |path | 用于标识操作实体所属 Plugin。 | 
| EntityId | string | true | path/query | 实体的 Id。`plugins/abcd/entities/test123 或 plugins/abcd/entities?id=test123`。|
| Type | string | true | header/query | 用于标识实体的类型。|
| Source | string | true | header/query | 用于标识请求的发起 Plugin。|
| Owner | string | true | header/query | 用于标识请求的发起用户。|

```bash
curl -X DELETE "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123" \
  -H "Source: abcd" \
  -H "Owner: admin" \
  -H "Type: DEVICE" 
```

### 筛选 Entities

- Method: **GET**
- URL:

```
http://localhost:3500/v1.0/invoke/core/method/plugins/{plugin}/entities?owner={owner}&type={type}
```

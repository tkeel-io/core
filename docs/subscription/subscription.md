# 订阅

----
每个plugin在注册的时候自动创建一个与core交互的pubsub,名称为plugin的名称。 topic统一为pubcore，subcore，只有core与该plugin有相关权限
比如
pluginA: pluginA。pubcore用于向core发布数据，subcore用于从core订阅数据
## 订阅分类

1. 实时订阅（收到消息就触发）
2. 变更订阅（属性有变更时触发）
3. 周期订阅（周期性上报所有属性）

## 订阅的实现
1. 筛选数据
2. 数据计算和变换
3. 发送数据

## 订阅的表达形式

```json
{
  "source": "pluginA",
  "filter": "/abcd/+",
  "target": "pluginB",
  "mode": "realtime"
}
```
```json
{
  "source": "pluginA",
  "filter": "* where thing_id=abcd",
  "target": "pluginA",
  "mode": "realtime"
}
```
其中filter可采用不同的表达形式


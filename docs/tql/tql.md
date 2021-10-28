## 简介

`TQL`即Tkeel QL， 主要用于数据选择，如下
1. core 与各个插件之间的订阅的数据选择 
2. core 内部entity之间的映射（mapper）的数据选择 


提供的功能：

1. `tql`的解析，提供 API 将 `tql` 语句字符串转化为 TargetEntity，SourceActors，Tentacles。
2.  计算结果，提供 API 输入参数 按照`tql`语法计算出结果并输出。 （订阅不支持这一步， 仅mapper）



## Demo1


> 现有三个实体，将entity1和entity2的部分数据映射到entity3。
```
TQL:
    insert into entity3 select
		entity1.property1 as property1,
		entity2.property2.name as property2,
		entity1.property1 + entity2.property3 as property3

```
`insert into` 必填，

`entity3` 订阅可以是订阅的ID， mapper是entity ID 

`select` 必填， 后面支持通配符

`as` 可选， 订阅没有as， mapper有as


> 1. `tql`解析， 输出如下`json`。
```json
{
  "TargetEntity": "entity3",
  "SourceEntities": ["entity1", "entity2"],
  "Tentacles": {
    "entity1": ["property1"],
    "entity2": ["property2", "property3"]
  }
}
```


> 2. 计算结果,输入为Input(map), 输出为Output(map)。
```json
Input:
{
  "entity1.property1": 1,
  "entity2.property2.name": 2,
  "entity2.property3": 3
}

Output:
{
  "property1": 1,
  "property2": 2,
  "property3": 4
}
```
后续再考虑支持获取单步计算结果。



## Demo2

场景：`设备管理` 订阅所有的 `设备接入` 数据, 
```
TQL:
    insert into sub_entity_id select *

```
1. tql解析输出
```json
{
  "TargetEntity": "sub_entity_id",
  "SourceEntities": ["*"],
  "Tentacles": {
    "*": ["*"]
  }
}
```

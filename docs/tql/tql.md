## 简介

`TQL`即Tkeel QL， 主要用于数据选择，如下
1. core 与各个插件之间的订阅的数据选择 
2. core 内部entity之间的映射（mapper）的数据选择 


提供的功能：

1. `tql`的解析，提供 API 将 `tql` 语句字符串转化为 TargetEntity，SourceActors，Tentacles。
2.  计算结果，提供 API 输入参数 按照`tql`语法计算出结果并输出。



## Demo


> 现有三个实体，将entity1和entity2的部分数据映射到entity3。
```
TQL:
    insert into entity3 select
		entity1.property1 as property1,
		entity2.property2.name as property2,
		entity1.property1 + entity2.property3 as property3

```
`insert into` 必填
`entity3` entity 
`as` 可选， 订阅没有， mapper有


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





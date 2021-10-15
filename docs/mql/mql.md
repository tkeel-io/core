## 简介

`MQL`即Mapper QL， 用于针对海量`Actor`的属性数据的映射。`Mapper`的定义请查[mapper.md](../mapper/mapper.md)，`MQL`所做的事就是将一个`json`输入转换成另一个`json`输出。


处理`MQL`：

1. `mql`的静态解析，得到TargetEntity，SourceActors，Tentacles。
2.  运行时执行`json`的转换。





## Demo


> 现有三个实体，将entity1和entity2的部分数据映射到entity3。
```sql
MQL:

		entity1.property1 as property1,
		entity2.property2 as property2,
		entity1.property3 + entity2.property3 as property3

```


> 解析`MQL`然后可以得到`TargetEntity`，`SourceEntities`，`Tentacles`。
```bash
Parse:
	TargetEntity() returns:
		entity3
	SourceEntities() returns:
		entity1, entity2
    Tentacles:
        {
            "entity1": ["property1", "property3"],
            "entity2": ["property2", "property3"]
        }
```


> 执行`MQL`，以`json`作为输入，输出`json`作为entity3的映射。
```json
Input:
	{
        "entity3": {
			"property1": 0,
			"property2": 0,
			"property3": 0
		},
		"entity1": {
			"property1": 12,
			"property2": "say hello.",
			"property3": 50
		},
		"entity2": {
			"property1": "i am entity2.",
			"property2": 22,
			"property3": 33
		}
	}

Output:
	{
        "entity3": {
			"property1": 12,
			"property2": 22,
			"property3": 83
		},
		"entity1": {
			"property1": 12,
			"property2": "say hello.",
			"property3": 50
		},
		"entity2": {
			"property1": "i am entity2.",
			"property2": 22,
			"property3": 33
		}
	}
```





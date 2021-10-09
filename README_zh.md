# Keel
[English](README.md)

![img.png](docs/images/img/architecture.png)

TKeel 解决了构建高性能、模块化数据接入平台的关键要求。 它利用微服务架构模式并提供可拔插架构以及高速数据平面，帮助您快速构建健壮可复用的物联网解决方案。

## How it works

![img.png](docs/images/img/layer.png)

 - Core 代表了一种模式，它包含一些数据组织形式以及处理方式。
    - Core 通过时序数据、属性数据、关系数据来构建不同的对象。节点的唯一性由 ID 来保证
    - 通过 快照+订阅（Event数据） 来解决数据交换。
    
    

## 实体

## 关系
关系是实体与实体之间的关系（包括自己与自己之间的关系）


	// select sum(light1.a, light2.a) as house.a
	// select light1.b as house.b
	// select sum(light1.a, light1.b) as house.d
	// select house.a as house.c
	// select sum(2*light1.a, light2.a) as house.e


1. 简单映射
    ```sql
    select light1.b as house.b
    ``` 
2. 计算+映射
    ```sql
    select sum(light1.a, light2.a) as house.a
    ```
3. 多对一映射+计算
    ```sql
   	select sum(2*light1.a, light2.a) as house.e
	
    ```
4. 自身映射+计算
    ```sql
	select sum(light1.a, light2.a) as house.a
    ```
    
 
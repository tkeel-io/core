# Core
[English](README.md)

![img.png](docs/images/architecture.png)

Core通过api对外提供属性搜索，时序查询，数据写入，数据查询，数据订阅等服务
    
## 实体
物联网世界里的操作对象，以及这些对象组合抽象出来的对象，包括网关，设备，设备的聚合抽象等等。  
实体具有属性，属性对一个实体某种信息的描述  

## actor
实体在程序中的存在形式，内存中存储了自己的属性，以及其他实体映射过来的属性

## 关系
关系是实体与实体之间的关系

## 映射
映射的操作包含两个部分: 写复制和计算更新
![img.png](docs/images/mapping.png)
### 映射（写复制）
自身属性的变更可能触发写复制到其他实体

### 映射（计算+更新）
其他实体映射过来的属性的变更，可能触发计算和更新自身属性


1. 简单映射
    ```sql
    select light1.a as house.a
    ``` 
2. 计算+映射
    ```sql
    select sum(light1.b, light2.b) as house.b
    ```
3. 多对一映射+计算
    ```sql
   	select sum(2*light1.a, light2.a) as house.e
    ```
4. 自身映射+计算
    ```sql
	select sum(light1.c, light1.d) as light1.e
    ```

## 数据的传递
![img.png](docs/images/message_passing.png)
 
 蓝色线条代表上行，黑色代表下行
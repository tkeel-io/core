# 架构说明

----

## 前后端分离

tKeel 将 [前端](https://github.com/xujielong/console) 与 [后端](https://github.com/xujielong/demo) 分开，实现了面向云原生的设计，后端的各个功能组件可通过 REST API 对接外部系统。 可参考 [API文档](docs/api/index.md)。
下图是系统架构图。 tKeel 无底层的基础设施依赖，可以运行在私有云、公有云、VM 或物理环境（BM）之上。

![Architecture](docs/images/architecture.png)


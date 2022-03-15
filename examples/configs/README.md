## 分布式测试用例


```bash
# cwd: core/

# 使用 examples/configs/core0/config.yml 启动 core0
go run cmd/core/main.go --conf examples/configs/core0/config.yml
# 使用 examples/configs/core1/config.yml 启动 core1
go run cmd/core/main.go --conf examples/configs/core1/config.yml --http_addr=":6790" --grpc_addr=":31235"
```



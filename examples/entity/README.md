## run example

### 运行core
通过命令行运行core，core内部端口为6789，会定义一个core-pubsub自己使用，一个client-pubsub供client发送数据和一个client-subscription供core消费client发送的数据使用。
```bash
dapr run --app-id core --app-protocol http --app-port 6789 --dapr-http-port 3500 --dapr-grpc-port 50001 --log-level debug  --components-path ./examples/configs/core  go run . serve
```
```bash
INFO[0001] application discovered on port 6789           app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
INFO[0001] application configuration loaded              app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
INFO[0001] actor runtime started. actor idle timeout: 1h0m0s. actor scan interval: 30s  app_id=core instance=i-i2j8ujhr scope=dapr.runtime.actor type=log ver=1.4.2
DEBU[0001] try to connect to placement service: dns:///localhost:50005  app_id=core instance=i-i2j8ujhr scope=dapr.runtime.actor.internal.placement type=log ver=1.4.2
DEBU[0001] app responded with subscriptions [{core-pubsub core-pub map[] [0xc000bcbb40] []}]  app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
INFO[0001] app is subscribed to the following topics: [core-pub] through pubsub=core-pubsub  app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
INFO[0001] app is subscribed to the following topics: [core-pub] through pubsub=client-pubsub  app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
DEBU[0001] subscribing to topic=core-pub on pubsub=client-pubsub  app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
DEBU[0001] subscribing to topic=core-pub on pubsub=core-pubsub  app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
INFO[0001] dapr initialized. Status: Running. Init Elapsed 1335.3118729999999ms  app_id=core instance=i-i2j8ujhr scope=dapr.runtime type=log ver=1.4.2
DEBU[0001] established connection to placement service at dns:///localhost:50005  app_id=core instance=i-i2j8ujhr scope=dapr.runtime.actor.internal.placement type=log ver=1.4.2
DEBU[0001] placement order received: lock                app_id=core instance=i-i2j8ujhr scope=dapr.runtime.actor.internal.placement type=log ver=1.4.2
DEBU[0001] placement order received: update              app_id=core instance=i-i2j8ujhr scope=dapr.runtime.actor.internal.placement type=log ver=1.4.2
INFO[0001] placement tables updated, version: 8          app_id=core instance=i-i2j8ujhr scope=dapr.runtime.actor.internal.placement type=log ver=1.4.2
DEBU[0001] placement order received: unlock              app_id=core instance=i-i2j8ujhr scope=dapr.runtime.actor.internal.placement type=log ver=1.4.2
```

### 运行client
client通过api创建一个实体，通过pubsub更新实体的属性，然后获取当前实体的快照。[code](examples/enitity/main.go)
```bash
cd examples/entity
dapr run --app-id client  --log-level debug  --components-path ../configs/entity/ go run .
```
```bash
ℹ️  Checking if Dapr sidecar is listening on GRPC port 28007
ℹ️  Dapr sidecar is up and running.
ℹ️  Updating metadata for app command: go run .
✅  You're up and running! Both Dapr and your app logs will appear here.

== APP == dapr client initializing for: 127.0.0.1:28007
DEBU[0002] no mDNS address found in cache, browsing network for app id core  app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] Browsing for first mDNS address for app id core  app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] mDNS response for app id core received.       app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] Adding IPv4 address 192.168.181.2:33773 for app id core cache entry.  app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] mDNS browse for app id core canceled.         app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] Browsing for first mDNS address for app id core canceled.  app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] Refreshing mDNS addresses for app id core.    app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] mDNS response for app id core received.       app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
DEBU[0002] Adding IPv4 address 192.168.181.2:33773 for app id core cache entry.  app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
== APP == {"id":"test1","type":"device","owner":"abc","status":"active","version":1,"plugin_id":"pluginA","last_time":1634780763746,"properties":{}}
DEBU[0002] found mDNS IPv4 address in cache: 192.168.181.2:33773  app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
== APP == {"id":"test1","type":"device","owner":"abc","status":"active","version":1,"plugin_id":"pluginA","last_time":1634780763746,"properties":{}}
DEBU[0002] found mDNS IPv4 address in cache: 192.168.181.2:33773  app_id=client instance=i-i2j8ujhr scope=dapr.contrib type=log ver=1.4.2
== APP == {"id":"test1","type":"device","owner":"abc","status":"active","version":1,"plugin_id":"pluginA","last_time":1634780763750,"properties":{"core":{"time":1634780763749,"value":189}}}
✅  Exited App successfully
ℹ️
terminated signal received: shutting down
```





### publish Event

```bash
curl -X POST http://localhost:3500/v1.0/publish/core-pubsub/core-pub \
  -H "Content-Type: application/json" \
  -d '{
       "entity_id": "test234",
       "owner": "admin",
       "plugin": "abcd",
       "data": {
           "temp": 234
       }
     }'

```

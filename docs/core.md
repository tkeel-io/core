#### 1  时序图
##### 1.1 Search Service
1. Search
```puml
@startuml

header Search

actor Client
participant SearchService
participant SearchDriver
participant Elasticsearch


client->SearchService:request
SearchService->SearchDriver:search
SearchDriver->Elasticsearch:call
Elasticsearch-->SearchDriver:return
SearchDriver-->SearchService:return
SearchService-->client:response

@enduml
```
2. DeleteByID
```puml
@startuml

header DeleteByID

actor Client
participant SearchService
participant SearchDriver
participant Elasticsearch

client->SearchService:request
SearchService->SearchDriver:Delete
SearchDriver->Elasticsearch:call
Elasticsearch-->SearchDriver:return
SearchDriver-->SearchService:return
SearchService-->client:response

@enduml
```
3. Index
```puml
@startuml

header BuildIndex

actor Client
participant SearchService
participant SearchDriver
participant Elasticsearch

client->SearchService:request
SearchService->SearchDriver:BuildIndex
SearchDriver->Elasticsearch:call
Elasticsearch-->SearchDriver:return
SearchDriver-->SearchService:return
SearchService-->client:response

@enduml
```

##### 1.2 EntityService
1. CreateEntity
```puml
@startuml

header CreateEntity

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: CreateEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETSystem 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService->EntityService:create mapper
EntityService-->client:response
@enduml
```
2. UpdateEntity
```puml
@startuml

header UpdateEntity

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService->EntityService:create mapper
EntityService-->client:response

@enduml
```
3. GetEntity
```puml
@startuml

header GetEntity

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: GetEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@endeml
```
4. DeleteEntity
```puml
@startuml

header DeleteEntity

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: DeleteEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETSystem 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
5. UpdateEntityProps
````puml
@startuml

header UpdateEntityProps

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
````
6. PatchEntityProps
```puml

header PatchEntityProps

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

```
7. PatchEntityPropsZ
```puml
@startuml

header PatchEntityPropsZ

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
8. RemoveEntityProps
```puml
@startuml

header RemoveEntityProps

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
9. UpdateEntityConfigs
```puml
@startuml

header UpdateEntityConfigs

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
10. PatchEntityConfigs
```puml
@startuml

header PatchEntityConfigs

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
11. GetEntityConfigs
```puml
@startuml

header GetEntityConfigs

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
12. RemoveEntityConfigs
```puml
@startuml

header RemoveEntityConfigs

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
13. ListEntity
```puml
@startuml

header ListEntity

actor  client
participant EntityService
participant ProxyService
participant apiManager
participant holder
participant Elasticsearch
control dispatcher
queue kafka
participant runtime

client->EntityService:request
EntityService->apiManager:call: PatchEntity
apiManager->holder:await response
apiManager->dispatcher:msg router
dispatcher->kafka:msg pub
kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:dispatch:EventType:v1.ETCallback
dispatcher->ProxyService:/v1/respond
ProxyService->apiManager:OnRespond
apiManager->holder:OnRespond
holder-->apiManager:response
apiManager-->EntityService
EntityService-->client:response

@enduml
```
##### 1.3 SubscriptionService
1. CreateSubscription
```puml
@startuml

header CreateSubscription

actor client
participant SubService
participant apiManager
participant EntityRepo
participant EtcdDao


client->SubService:request
SubService->apiManager:call: CreateSubscription
apiManager->EntityRepo:call: PutSubscription
EntityRepo->EtcdDao:call: PutResource
SubService-->client:response

@enduml
```
2. UpdateSubscription
```puml
@startuml

header UpdateSubscription

actor client
participant SubService
participant apiManager
participant EntityRepo
participant EtcdDao


client->SubService:request
SubService->apiManager:call: CreateSubscription
apiManager->EntityRepo:call: PutSubscription
EntityRepo->EtcdDao:call: PutResource
SubService-->client:response

@enduml
```
3. DeleteSubscription
```puml
@startuml

header DeleteSubscription

actor client
participant SubService
participant apiManager
participant EntityRepo
participant EtcdDao


client->SubService:request
SubService->apiManager:call: GetSubscription
apiManager->EntityRepo:call: DelSubscription
EntityRepo->EtcdDao:call: GetResource
EtcdDao-->SubService:return: sub info

SubService->apiManager:call: DeleteSubscription
apiManager->EntityRepo:call: DelSubscription
EntityRepo->EtcdDao:call: DelResource
SubService-->client:response

@enduml
```
4. GetSubscription
5. ListSubscription

##### 1.4 TopicService
1.TopicEventHandler
```puml
@startuml

header TopicEventHandler

actor client
participant TopicService
control dispatcher
participant Consumer
queue       kafka
participant runtime

client->TopicService:request
TopicService->dispatcher:msg router:EventType:v1.ETEntity
dispatcher->kafka:msg producer
TopicService-->client:response

kafka->runtime:HandleMessage
runtime->runtime:DeliveredEvent:Unmarshal
runtime->runtime:HandleEvent
runtime->runtime:PrepareEvent:EventType:v1.ETEntity 
runtime->runtime:prepareSystemEvent
runtime->runtime:handleCallback
runtime->dispatcher:DispatchToLog
@enduml
```
##### 1.5 TsService
1. GetTSData
```puml
@startuml

header GetTSData

actor client
participant TsService
participant tseriesClient
participant  clickhouse

client->TsService
TsService->tseriesClient
tseriesClient->clickhouse:call: Query
clickhouse->TsService
TsService->client:response
@enduml
```
##### 1.7 RawDataService
1. GetRawdata
```puml
@startuml

header GetRawdata

actor client
participant RawDataService
participant tseriesClient
participant  clickhouse

client->TsService
TsService->tseriesClient
tseriesClient->clickhouse:call: Query
clickhouse->TsService
TsService->client:response
@enduml
```
##### 1.8 GOPSService
1. Metrics
2. Debug
3. SetNode
##### 1.9 MetricsService
1. Metrics

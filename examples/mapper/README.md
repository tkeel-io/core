## Mapper Example

## run core

```bash 
cd core/
# run...
dapr run --app-id core --app-protocol http --app-port 6789 --dapr-http-port 3500 --dapr-grpc-port 50001 --log-level debug  --components-path ./examples/configs/core  go run . serve
```



## create entities

```bash
# create entity through pubsub-event.
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

# query test234
curl -X GET "http://localhost:3500/v1.0/invoke/core/method/v1/plugins/abcd/entities/test123?source=abcd&owner=admin&type=DEVICE" 

# create test123 through APIs.
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/v1/plugins/abcd/entities?id=test123&source=abcd&owner=admin&type=DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
        "temp": 123
     }'



curl -X POST "http://localhost:3500/v1.0/invoke/core/method/v1/plugins/abcd/entities?id=test123&source=abcd&owner=admin&type=DEVICE" \
  -H "Content-Type: application/json" \
  -d '{
        "property1": {
          "test1": 123
        }
     }'

# update entity<test123>
curl -X PUT "http://localhost:3500/v1.0/invoke/core/method/v1/plugins/abcd/entities/test123" \
  -H "Content-Type: application/json" \
  -d '{
        "type": "DEVICE",
        "source": "abcd",
        "owner": "admin",
        "properties": {
          "temp": 123,
          "color": "red"
        }
     }'



# create mapper for test123.
curl -X PUT "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123/mappers?source=abcd&type=DEVICE&owner=admin" \
  -H "Content-Type: application/json" \
  -d '{
       "name": "subscribe-test234",
       "tql": "insert into test123 select test234.temp as temp"
     }'

# publish event for test234
curl -X POST http://localhost:3500/v1.0/publish/core-pubsub/core-pub \
  -H "Content-Type: application/json" \
  -d '{
       "entity_id": "test123",
       "owner": "admin",
       "plugin": "abcd",
       "data": {
           "temp": 1233
       }
     }'

# query test123
curl -X GET "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/entities/test123" \
  -H "Source: abcd" \
  -H "Owner: admin"  \
  -H "Type: DEVICE"


# update entity config.
curl -X PUT "http://localhost:3500/v1.0/invoke/core/method/v1/plugins/abcd/entities/test123/configs?source=abcd&type=DEVICE&owner=admin" \
  -H "Content-Type: application/json" \
  -d '[
          {
            "id": "property1",
            "type": "int",
            "define": {},
            "enabled": true,
            "enabled_search": true
          }
    ]'



curl -X PUT "http://localhost:3500/v1.0/invoke/core/method/v1/plugins/abcd/entities/test123/configs?source=abcd&type=DEVICE&owner=admin" \
  -H "Content-Type: application/json" \
  -d '[
          {
            "id": "property1",
            "type": "struct",
            "define": {
              "fields": [
                {
                  "id": "test1",
                  "type": "int",
                  "enabled": true,
                  "enabled_search": true,
                  "define": {}
                }
              ]
            },
            "enabled": true,
            "enabled_search": true
          }
    ]'

```






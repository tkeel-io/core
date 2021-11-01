## Subscription APIs

> Subscription 本质也是 Entity 的一种，不过我们为了更简单的操作 Subscription，将其APIs独立出来。



### Subscription Get
```bash
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/subscriptions/sub123?owner=admin&type=SUBSCRIPTION" \
  -H "Content-Type: application/json" \
  -H "Source: abcd" 
```

### Subscription Create
```bash
curl -X POST "http://localhost:3500/v1.0/invoke/core/method/plugins/abcd/subscriptions?id=sub123&owner=admin&type=SUBSCRIPTION" \
  -H "Content-Type: application/json" \
  -H "Source: abcd" \
  -d '{
        "mode": "realtime",
        "source": "ignore",
        "filter":"insert into sub123 select test123.temp",
        "target": "ignore",
        "topic": "sub123",
        "pubsub_name": "core-pubsub"
     }'
```

### Subscription Update
```bash
put .../plugins/{plugin}/subscriptions/{subscription}
```

### Subscription Delete
```bash
delete .../plugins/{plugin}/subscriptions/{subscription}
```

### Subscription List
```bash
get .../plugins/{plugin}/subscriptions
```

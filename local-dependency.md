## influxdb

```bash
sudo docker run -d -p 8086:8086 --name=influxdb -v /home/user/influxdb:/var/lib/influxdb influxdb
```


```bash
# write data via API-v2
curl --request POST "http://localhost:8086/api/v2/write?org=yunify&bucket=entity&precision=s" \
--header "Authorization: Token 9bUWcVwUpxbNSuhJMLbRaJxCVl8LzFV33znGx-pAXg4HUxFgWRTkRArF5Z9lMDcOn1pzzfD4dovLkkTnxuVMtg==" \
--data-raw "
mem,host=host1 used_percent=23.43234543 1640077883
mem,host=host2 used_percent=26.81522361 1640077883
mem,host=host1 used_percent=22.52984738 1640077883
mem,host=host2 used_percent=27.18294630 1640077883
"

curl --request POST \
http://localhost:8086/api/v2/query?org=yunify  \
--header 'Authorization: Token 9bUWcVwUpxbNSuhJMLbRaJxCVl8LzFV33znGx-pAXg4HUxFgWRTkRArF5Z9lMDcOn1pzzfD4dovLkkTnxuVMtg==' \
--header 'Accept: application/csv' \
--header 'Content-type: application/vnd.flux' \
--data 'from(bucket:"entity")
      |> range(start: -12h)'
```
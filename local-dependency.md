## influxdb

```bash
sudo docker run -p 8083:8083 -p8086:8086 --expose 8090 --expose 8099 --name influxdb\
      -e DOCKER_INFLUXDB_INIT_MODE=upgrade \
      -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
      -e DOCKER_INFLUXDB_INIT_PASSWORD=admin123 \
      -e DOCKER_INFLUXDB_INIT_ORG=org123 \
      -e DOCKER_INFLUXDB_INIT_BUCKET=bucket123 \
      influxdb:2.0
```
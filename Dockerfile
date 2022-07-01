############################################################ 
# Dockerfile to build golang Installed Containers 

# Based on alpine

############################################################

FROM golang:1.17 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY="https://goproxy.cn,direct" make build

FROM alpine:3.13
RUN apk update
RUN apk add tzdata

RUN mkdir /keel
COPY --from=builder /src/dist/linux_amd64/release/core /keel


EXPOSE 6789
WORKDIR /keel
CMD ["/keel/core", "--search_engine", \
    "es://admin:admin@tkeel-middleware-elasticsearch-master:9200", \
    "--etcd", "http://tkeel-middleware-etcd:2379", \
    "--conf", "/config/config.yml"]

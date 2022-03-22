############################################################ 
# Dockerfile to build golang Installed Containers 

# Based on alpine

############################################################

FROM golang:1.17 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build

FROM alpine:3.13

RUN mkdir /keel
COPY --from=builder /src/dist/linux_amd64/release/core /keel


EXPOSE 6789
WORKDIR /keel
CMD ["/keel/core", "--search_engine", \
    "es://admin:admin@elasticsearch-master:9200", \
    "--etcd", "http://etcd:2379", \
    "--conf", "/config/config.yml"]

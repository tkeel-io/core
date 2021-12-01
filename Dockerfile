############################################################ 
# Dockerfile to build golang Installed Containers 
 
# Based on alpine
 
############################################################

FROM golang:1.16 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build

FROM alpine:3.13

RUN mkdir /keel
COPY --from=builder /src/dist/linux_amd64/release/core /keel

EXPOSE 6789
WORKDIR /keel
CMD ["/keel/core", "--search_brokers", "http://elasticsearch-master:9200"]

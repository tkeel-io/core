module github.com/tkeel-io/core

go 1.16

require (
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20211026222012-6af4c774c47b
	github.com/dapr/dapr v1.5.1 // indirect
	github.com/dapr/go-sdk v1.3.0
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/fsnotify/fsnotify v1.5.1
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0
	github.com/influxdata/influxdb-client-go v1.4.0
	github.com/mitchellh/mapstructure v1.4.2
	github.com/olivere/elastic/v7 v7.0.29
	github.com/panjf2000/ants/v2 v2.4.6
	github.com/pkg/errors v0.9.1
	github.com/shamaton/msgpack/v2 v2.1.0
	github.com/smartystreets/assertions v1.2.0
	github.com/smartystreets/gunit v1.4.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/tkeel-io/collectjs v0.0.0-20211130035606-e8d64c4a2a39
	github.com/tkeel-io/kit v0.0.0-20211122161648-31b3986683f2
	github.com/tkeel-io/tkeel-interface/openapi v0.0.0-20211201125403-d4d4343c7730
	go.etcd.io/etcd/api/v3 v3.5.1
	go.etcd.io/etcd/client/v3 v3.5.1
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20211216030914-fe4d6282115f
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/antlr/antlr4 => github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20211221011931-643d94fcab96

# MDMP Makefile

GOCMD = GO111MODULE=on go


GOTEST = $(GOCMD) test
BINNAME = core

export GO111MODULE ?= on
export GOPROXY ?= https://proxy.golang.org
export GOSUMDB ?= sum.golang.org
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty)
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_BRANCH=$(shell git name-rev --name-only HEAD)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
CGO			?= 0
CLI_BINARY  = core

ifdef REL_VERSION
	CORE_VERSION := $(REL_VERSION)
else
	CORE_VERSION := edge
endif

ifdef API_VERSION
	RUNTIME_API_VERSION = $(API_VERSION)
else
	RUNTIME_API_VERSION = 1.0
endif

LOCAL_ARCH := $(shell uname -m)
ifeq ($(LOCAL_ARCH),x86_64)
	TARGET_ARCH_LOCAL = amd64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),armv8)
	TARGET_ARCH_LOCAL = arm64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),aarch64)
	TARGET_ARCH_LOCAL = arm64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 4),armv)
	TARGET_ARCH_LOCAL = arm
else
	TARGET_ARCH_LOCAL = amd64
endif
export GOARCH ?= $(TARGET_ARCH_LOCAL)

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
   TARGET_OS_LOCAL = linux
   GOLANGCI_LINT:=golangci-lint
   export ARCHIVE_EXT = .tar.gz
else ifeq ($(LOCAL_OS),Darwin)
   TARGET_OS_LOCAL = darwin
   GOLANGCI_LINT:=golangci-lint
   export ARCHIVE_EXT = .tar.gz
else
   TARGET_OS_LOCAL ?= windows
   BINARY_EXT_LOCAL = .exe
   GOLANGCI_LINT:=golangci-lint.exe
   export ARCHIVE_EXT = .zip
endif
export GOOS ?= $(TARGET_OS_LOCAL)
export BINARY_EXT ?= $(BINARY_EXT_LOCAL)

ifeq ($(origin DEBUG), undefined)
  BUILDTYPE_DIR:=release
else ifeq ($(DEBUG),0)
  BUILDTYPE_DIR:=release
else
  BUILDTYPE_DIR:=debug
  GCFLAGS:=-gcflags="all=-N -l"
  $(info $(H) Build with debugger information)
endif

run:
	@echo "---------------------------"
	@echo "-         Run             -"
	@echo "---------------------------"
	@$(GOCMD) run cmd/core/main.go serve

drun:
	dapr run --app-id core --app-protocol http --app-port 6789 --dapr-http-port 3500 --dapr-grpc-port 50001 --log-level debug  --components-path ./examples/configs/core  dlv debug ./dist/linux_amd64/release/core -- --conf ./config.yml



################################################################################
# Go build details                                                             #
################################################################################
BASE_PACKAGE_NAME := github.com/tkeel-io/core
OUT_DIR := ./dist

BINS_OUT_DIR := $(OUT_DIR)/$(GOOS)_$(GOARCH)/$(BUILDTYPE_DIR)
LDFLAGS :="-X $(BASE_PACKAGE_NAME)/pkg/version.GitCommit=$(GIT_COMMIT) -X $(BASE_PACKAGE_NAME)/pkg/version.GitBranch=$(GIT_BRANCH) -X $(BASE_PACKAGE_NAME)/pkg/version.GitVersion=$(GIT_VERSION) -X $(BASE_PACKAGE_NAME)/pkg/version.BuildDate=$(BUILD_DATE) -X $(BASE_PACKAGE_NAME)/pkg/version.Version=$(CORE_VERSION)"

INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
API_PROTO_FILES := api/core/v1/entity.proto api/core/v1/subscription.proto api/core/v1/list.proto api/core/v1/search.proto api/core/v1/ts.proto api/core/v1/topic.proto api/core/v1/event.proto

.PHONY: init
# init env
init:
	go get -d -u  github.com/tkeel-io/tkeel-interface/openapi@latest
	go get -d -u  github.com/tkeel-io/kit@latest
	go get -d -u  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.0

	go install  github.com/tkeel-io/tkeel-interface/tool/cmd/artisan@latest
	go install  google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go install  google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go install  github.com/tkeel-io/tkeel-interface/protoc-gen-go-http@latest
	go install  github.com/tkeel-io/tkeel-interface/protoc-gen-go-errors@latest
	go install  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.0

.PHONY: api
# generate api proto
api:
	protoc --proto_path=. \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:. \
 	       --go-http_out=paths=source_relative:. \
 	       --go-grpc_out=paths=source_relative:. \
 	       --go-errors_out=paths=source_relative:. \
 	       --openapiv2_out=./api/ \
		   --openapiv2_opt=allow_merge=true \
 	       --openapiv2_opt=logtostderr=true \
 	       --openapiv2_opt=json_names_for_fields=false \
	       $(API_PROTO_FILES)

	@echo "---------------------------------------------------------"
	@echo "----- 请注意 core/api/core/v1/topic_http.pb.go 的变更 -----"
	@echo "---------------------------------------------------------"


.PHONY: api-docs
# generate api docs
api-docs:
	artisan markdown -f api/apidocs.swagger.json  -t third_party/markdown-templates/ -o docs/APIs/Core -m all



build:
	@echo "---------------------------"
	@echo "-        build...         -"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GCFLAGS) -ldflags $(LDFLAGS) \
     	-o $(BINS_OUT_DIR)/$(CLI_BINARY)$(BINARY_EXT) cmd/core/main.go
	@echo "-    builds completed!    -"
	@echo "---------------------------"
	@echo "Bin: $(BINS_OUT_DIR)/$(CLI_BINARY)$(BINARY_EXT)"
	@echo "-----------Done------------"

show:
	@$(BINS_OUT_DIR)/$(CLI_BINARY)$(BINARY_EXT) --version

ifeq ($(GOOS),windows)
BINARY_EXT_LOCAL:=.exe
GOLANGCI_LINT:=golangci-lint.exe
export ARCHIVE_EXT = .zip
else
BINARY_EXT_LOCAL:=
GOLANGCI_LINT:=golangci-lint
export ARCHIVE_EXT = .tar.gz
endif

################################################################################
# Archive                                                            #
################################################################################
archive: archive-$(CLI_BINARY)$(ARCHIVE_EXT)

ifeq ($(GOOS),windows)
archive-$(CLI_BINARY).zip:
	7z.exe a -tzip "$(ARCHIVE_OUT_DIR)\\$(CLI_BINARY)_$(GOOS)_$(GOARCH)$(ARCHIVE_EXT)" "$(BINS_OUT_DIR)\\$(CLI_BINARY)$(BINARY_EXT)"
else
archive-$(CLI_BINARY).tar.gz:
	chmod +x $(BINS_OUT_DIR)/$(CLI_BINARY)$(BINARY_EXT)
	tar czf "$(ARCHIVE_OUT_DIR)/$(CLI_BINARY)_$(GOOS)_$(GOARCH)$(ARCHIVE_EXT)" -C "$(BINS_OUT_DIR)" "$(CLI_BINARY)$(BINARY_EXT)"
endif

test:
ifeq ($(GOOS), windows)
	@go test -v -cover -gcflags=all=-l .\...
else
	@go test -v -cover -gcflags=all=-l -coverprofile=coverage.out ./...
endif

docker-build:
	sudo docker build -t tkeelio/core:${CORE_VERSION} .
docker-push:
	sudo docker push tkeelio/core:${CORE_VERSION}

docker-auto:
	sudo docker build -t tkeelio/core:${CORE_VERSION} .
	sudo docker push tkeelio/core:${CORE_VERSION}


################################################################################
# Target: lint                                                                 #
################################################################################
# Due to https://github.com/golangci/golangci-lint/issues/580, we need to add --fix for windows
.PHONY: lint
lint:
	$(GOLANGCI_LINT) run --timeout=30m


.PHONY: install generate

-include .dev/*.makefile

################################################################################
# Target: release                                                              #
################################################################################
.PHONY: release
release: build archive

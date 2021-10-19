# MDMP Makefile

GOCMD = GO111MODULE=on go

VERSION := $(shell grep "const Version " cmd/root.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
GORUN = $(GOCMD) run -ldflags "-X github.com/tkeel-io/core/cmd.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/tkeel-io/core/cmd/cmd.BuildDate=${BUILD_DATE}"
GOBUILD = $(GOCMD) build -ldflags "-X github.com/tkeel-io/core/cmd.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/tkeel-io/core/cmd/cmd.BuildDate=${BUILD_DATE}"
GOTEST = $(GOCMD) test
BINNAME = core

export GO111MODULE ?= on
export GOPROXY ?= https://proxy.golang.org
export GOSUMDB ?= sum.golang.org
GIT_COMMIT  = $(shell git rev-list -1 HEAD)
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty)
CGO			?= 0
CLI_BINARY  = core

ifdef REL_VERSION
	CLI_VERSION := $(REL_VERSION)
else
	CLI_VERSION := edge
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
	@$(GORUN) . serve


################################################################################
# Go build details                                                             #
################################################################################
BASE_PACKAGE_NAME := github.com/tkeel-io/core
OUT_DIR := ./dist

BINS_OUT_DIR := $(OUT_DIR)/$(GOOS)_$(GOARCH)/$(BUILDTYPE_DIR)
LDFLAGS := "-X github.com/tkeel-io/core/cmd.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/tkeel-io/core/cmd/cmd.BuildDate=${BUILD_DATE}"

build:
	@echo "---------------------------"
	@echo "-        build...         -"
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GCFLAGS) -ldflags $(LDFLAGS) \
     	-o $(BINS_OUT_DIR)/$(CLI_BINARY)$(BINARY_EXT)
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

docker-build: build
	docker build -t tkeelio/core:0.0.1 .
docker-push:
	docker push tkeelio/core:0.0.1

################################################################################
# Target: lint                                                                 #
################################################################################
# Due to https://github.com/golangci/golangci-lint/issues/580, we need to add --fix for windows
.PHONY: lint
lint:
	$(GOLANGCI_LINT) run --timeout=20m


.PHONY: install generate

-include .dev/*.makefile

################################################################################
# Target: release                                                              #
################################################################################
.PHONY: release
release: build archive
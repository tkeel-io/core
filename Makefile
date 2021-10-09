# MDMP Makefile

GOCMD = GO111MODULE=on go

VERSION := $(shell grep "const Version " pkg/version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
GORUN = $(GOCMD) run -ldflags "-X github.com/tkeel-io/core/pkg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/tkeel-io/core/pkg/version.BuildDate=${BUILD_DATE}"
GOBUILD = $(GOCMD) build -ldflags "-X github.com/tkeel-io/core/pkg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/tkeel-io/core/pkg/version.BuildDate=${BUILD_DATE}"
GOTEST = $(GOCMD) test
BINNAME = core


run:
	@echo "---------------------------"
	@echo "-         Run             -"
	@echo "---------------------------"
	@$(GORUN) . serve

build:
	@rm -rf bin/
	@mkdir bin/
	@echo "---------------------------"
	@echo "-        build...         -"
	@$(GOBUILD)    -o bin/$(BINNAME)
	@echo "-     build(linux)...     -"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64  $(GOBUILD) -o bin/linux/$(BINNAME)
	@echo "-    builds completed!    -"
	@echo "---------------------------"
	@bin/$(BINNAME) version
	@echo "-----------Done------------"

ifeq ($(GOOS),windows)
BINARY_EXT_LOCAL:=.exe
GOLANGCI_LINT:=golangci-lint.exe
export ARCHIVE_EXT = .zip
else
BINARY_EXT_LOCAL:=
GOLANGCI_LINT:=golangci-lint
export ARCHIVE_EXT = .tar.gz
endif

test:

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

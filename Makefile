# MDMP Makefile

GOCMD = GO111MODULE=on go

VERSION := $(shell grep "const Version " pkg/version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
GORUN = $(GOCMD) run -ldflags "-X git.internal.yunify.com/tkeel/core/pkg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X git.internal.yunify.com/tkeel/core/pkg/version.BuildDate=${BUILD_DATE}"
GOBUILD = $(GOCMD) build -ldflags "-X git.internal.yunify.com/tkeel/core/pkg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X git.internal.yunify.com/tkeel/core/pkg/version.BuildDate=${BUILD_DATE}"
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

test:

docker-build: build
	docker build -t tkeelio/core:0.0.1 .
docker-push:
	docker push tkeelio/core:0.0.1

.PHONY: install generate

-include .dev/*.makefile
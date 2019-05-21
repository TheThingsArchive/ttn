SHELL = bash

# Environment

GIT_BRANCH = $(or $(CI_BUILD_REF_NAME) ,`git rev-parse --abbrev-ref HEAD 2>/dev/null`)
GIT_COMMIT = $(or $(CI_BUILD_REF), `git rev-parse HEAD 2>/dev/null`)
GIT_TAG = $(shell git describe --abbrev=0 --tags 2>/dev/null)
BUILD_DATE = $(or $(CI_BUILD_DATE), `date -u +%Y-%m-%dT%H:%M:%SZ`)
GO_PATH = $(shell echo $(GOPATH) | awk -F':' '{print $$1}')
PARENT_DIRECTORY= $(shell dirname $(PWD))
GO_SRC = $(shell pwd | xargs dirname | xargs dirname | xargs dirname)

# All

.PHONY: all build-deps deps dev-deps protos-clean protos protodoc mocks test cover-clean cover-deps cover coveralls fmt vet ttn ttnctl build link docs clean docker

all: deps build

# Deps

build-deps:

deps:
	go mod download

dev-deps: deps
	@command -v protoc-gen-grpc-gateway > /dev/null || go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	@command -v protoc-gen-gogottn > /dev/null || go install github.com/TheThingsNetwork/ttn/utils/protoc-gen-gogottn
	@command -v protoc-gen-ttndoc > /dev/null || go install github.com/TheThingsNetwork/ttn/utils/protoc-gen-ttndoc
	@command -v mockgen > /dev/null || go get github.com/golang/mock/mockgen
	@command -v golint > /dev/null || go get golang.org/x/lint/golint
	@command -v forego > /dev/null || go get github.com/ddollar/forego

dev-certs:
	ttn discovery gen-cert localhost 127.0.0.1 ::1 discovery --config ./.env/discovery/dev.yml
	ttn router gen-cert localhost 127.0.0.1 ::1 router --config ./.env/router/dev.yml
	ttn broker gen-cert localhost 127.0.0.1 ::1 broker --config ./.env/broker/dev.yml
	ttn networkserver gen-cert localhost 127.0.0.1 ::1 networkserver --config ./.env/networkserver/dev.yml
	ttn handler gen-cert localhost 127.0.0.1 ::1 handler --config ./.env/handler/dev.yml

# Go Test

GO_FILES = $(shell find . -name "*.go" | grep -vE ".git|.env|vendor|.pb.go|_mock.go")

GO_COVER_FILES = `find . -name "coverage.out"`

test: $(GO_FILES)
	go test ./...
	pushd api > /dev/null; go test ./...; popd > /dev/null
	pushd core/types > /dev/null; go test ./...; popd > /dev/null
	pushd mqtt > /dev/null; go test ./...; popd > /dev/null
	pushd utils/errors > /dev/null; go test ./...; popd > /dev/null
	pushd utils/random > /dev/null; go test ./...; popd > /dev/null
	pushd utils/security > /dev/null; go test ./...; popd > /dev/null
	pushd utils/testing > /dev/null; go test ./...; popd > /dev/null

cover-clean:
	rm -f $(GO_COVER_FILES) coverage.out.merged

cover-deps:
	@command -v goveralls > /dev/null || go get github.com/mattn/goveralls

cover: cover-clean $(GO_FILES)
	go test -coverprofile=coverage.out ./...
	pushd api > /dev/null; go test -coverprofile=coverage.out ./...; popd > /dev/null
	pushd core/types > /dev/null; go test -coverprofile=coverage.out ./...; popd > /dev/null
	pushd mqtt > /dev/null; go test -coverprofile=coverage.out ./...; popd > /dev/null
	pushd utils/errors > /dev/null; go test -coverprofile=coverage.out ./...; popd > /dev/null
	pushd utils/random > /dev/null; go test -coverprofile=coverage.out ./...; popd > /dev/null
	pushd utils/security > /dev/null; go test -coverprofile=coverage.out ./...; popd > /dev/null
	pushd utils/testing > /dev/null; go test -coverprofile=coverage.out ./...; popd > /dev/null
	echo "mode: set" > coverage.out.merged
	cat $(GO_COVER_FILES) | grep -vE "mode: set|/server.go|/manager_server.go" | sort > coverage.out.merged

coveralls: cover-deps cover
	goveralls -coverprofile=coverage.out.merged -service=travis-ci -repotoken $$COVERALLS_TOKEN

# Go Build

RELEASE_DIR ?= release
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOEXE = $(shell GOOS=$(GOOS) GOARCH=$(GOARCH) go env GOEXE)
CGO_ENABLED ?= 0

ifeq ($(GIT_BRANCH), $(GIT_TAG))
	TTN_VERSION = $(GIT_TAG)
	TAGS += prod
else
	TTN_VERSION = $(GIT_TAG)-dev
	TAGS += dev
endif

DIST_FLAGS ?= -a -installsuffix cgo

splitfilename = $(subst ., ,$(subst -, ,$(subst $(RELEASE_DIR)/,,$1)))
GOOSfromfilename = $(word 2, $(call splitfilename, $1))
GOARCHfromfilename = $(word 3, $(call splitfilename, $1))

GOVARS += -X main.version=${TTN_VERSION} -X main.gitBranch=${GIT_BRANCH} -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}
LDFLAGS = -ldflags "-w $(GOVARS)"
GOBUILD = CGO_ENABLED=$(CGO_ENABLED) GOOS=$(call GOOSfromfilename, $@) GOARCH=$(call GOARCHfromfilename, $@) go build $(DIST_FLAGS) ${LDFLAGS} -tags "${TAGS}" -o "$@"

ttn: $(RELEASE_DIR)/ttn-$(GOOS)-$(GOARCH)$(GOEXE)

$(RELEASE_DIR)/ttn-%: $(GO_FILES)
	$(GOBUILD) ./main.go

ttnctl: $(RELEASE_DIR)/ttnctl-$(GOOS)-$(GOARCH)$(GOEXE)

$(RELEASE_DIR)/ttnctl-%: $(GO_FILES)
	$(GOBUILD) ./ttnctl/main.go

build: ttn ttnctl

ttn-dev: DIST_FLAGS=
ttn-dev: CGO_ENABLED=1
ttn-dev: $(RELEASE_DIR)/ttn-$(GOOS)-$(GOARCH)$(GOEXE)

ttnctl-dev: DIST_FLAGS=
ttnctl-dev: CGO_ENABLED=1
ttnctl-dev: $(RELEASE_DIR)/ttnctl-$(GOOS)-$(GOARCH)$(GOEXE)

install:
	go install -v . ./ttnctl

dev: install ttn-dev ttnctl-dev

GOBIN ?= $(GO_PATH)/bin

link: build
	ln -sf $(PWD)/$(RELEASE_DIR)/ttn-$(GOOS)-$(GOARCH)$(GOEXE) $(GOBIN)/ttn
	ln -sf $(PWD)/$(RELEASE_DIR)/ttnctl-$(GOOS)-$(GOARCH)$(GOEXE) $(GOBIN)/ttnctl

# Documentation

docs:
	cd cmd/docs && HOME='$$HOME' go run generate.go > README.md
	cd ttnctl/cmd/docs && HOME='$$HOME' go run generate.go > README.md

# Clean

clean:
	[ -d $(RELEASE_DIR) ] && rm -rf $(RELEASE_DIR) || [ ! -d $(RELEASE_DIR) ]

docker: GOOS=linux
docker: GOARCH=amd64
docker: $(RELEASE_DIR)/ttn-linux-amd64
	docker build -t thethingsnetwork/ttn -f Dockerfile .

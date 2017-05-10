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
	@command -v govendor > /dev/null || go get "github.com/kardianos/govendor"

deps: build-deps
	govendor sync -v

dev-deps: deps
	@command -v protoc-gen-grpc-gateway > /dev/null || go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	@command -v protoc-gen-gogottn > /dev/null || go install github.com/TheThingsNetwork/ttn/utils/protoc-gen-gogottn
	@command -v protoc-gen-ttndoc > /dev/null || go install github.com/TheThingsNetwork/ttn/utils/protoc-gen-ttndoc
	@command -v mockgen > /dev/null || go get github.com/golang/mock/mockgen
	@command -v golint > /dev/null || go get github.com/golang/lint/golint
	@command -v forego > /dev/null || go get github.com/ddollar/forego

# Protobuf

PROTO_FILES = $(shell find api -name "*.proto" -and -not -name ".git")
COMPILED_PROTO_FILES = $(patsubst api%.proto, api%.pb.go, $(PROTO_FILES))
PROTOC_IMPORTS= -I/usr/local/include -I$(GO_PATH)/src -I$(PARENT_DIRECTORY) \
-I$(GO_PATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
PROTOC = protoc $(PROTOC_IMPORTS) \
--gogottn_out=plugins=grpc:$(GO_SRC) \
--grpc-gateway_out=:$(GO_SRC) `pwd`/

protos-clean:
	rm -f $(COMPILED_PROTO_FILES)

protos: $(COMPILED_PROTO_FILES)

api/%.pb.go: api/%.proto
	$(PROTOC)$<

protodoc: $(PROTO_FILES)
	protoc $(PROTOC_IMPORTS) --ttndoc_out=logtostderr=true,.lorawan.DevAddrManager=all:$(GO_SRC) `pwd`/api/protocol/lorawan/device_address.proto
	protoc $(PROTOC_IMPORTS) --ttndoc_out=logtostderr=true,.handler.ApplicationManager=all:$(GO_SRC) `pwd`/api/handler/handler.proto
	protoc $(PROTOC_IMPORTS) --ttndoc_out=logtostderr=true,.discovery.Discovery=all:$(GO_SRC) `pwd`/api/discovery/discovery.proto

# Mocks

mocks:
	mockgen -source=./api/protocol/lorawan/device.pb.go -package lorawan DeviceManagerClient > api/protocol/lorawan/device_mock.go
	mockgen -source=./api/networkserver/networkserver.pb.go -package networkserver NetworkServerClient > api/networkserver/networkserver_mock.go
	mockgen -source=./api/discovery/client.go -package discovery Client > api/discovery/client_mock.go

# Go Test

GO_FILES = $(shell find . -name "*.go" | grep -vE ".git|.env|vendor|.pb.go|_mock.go")
GO_PACKAGES = $(shell find . -name "*.go" | grep -vE ".git|.env|vendor" | sed 's:/[^/]*$$::' | sort | uniq)
GO_TEST_PACKAGES = $(shell find . -name "*_test.go" | grep -vE ".git|.env|vendor" | sed 's:/[^/]*$$::' | sort | uniq)
GO_COVER_PACKAGES = $(shell find . -name "*_test.go" | grep -vE ".git|.env|vendor|ttnctl|cmd|api" | sed 's:/[^/]*$$::' | sort | uniq)

GO_COVER_FILE ?= coverage.out
GO_COVER_DIR ?= .cover
GO_COVER_FILES = $(patsubst ./%, $(GO_COVER_DIR)/%.out, $(shell echo "$(GO_COVER_PACKAGES)"))

test: $(GO_FILES)
	go test $(GO_TEST_PACKAGES)

cover-clean:
	rm -rf $(GO_COVER_DIR) $(GO_COVER_FILE)

cover-deps:
	@command -v goveralls > /dev/null || go get github.com/mattn/goveralls

cover: $(GO_COVER_FILE)

$(GO_COVER_FILE): cover-clean $(GO_COVER_FILES)
	echo "mode: set" > $(GO_COVER_FILE)
	cat $(GO_COVER_FILES) | grep -vE "mode: set|/server.go|/manager_server.go" | sort >> $(GO_COVER_FILE)

$(GO_COVER_DIR)/%.out: %
	@mkdir -p "$(GO_COVER_DIR)/$<"
	go test -cover -coverprofile="$@" "./$<"

coveralls: cover-deps $(GO_COVER_FILE)
	goveralls -coverprofile=$(GO_COVER_FILE) -service=travis-ci -repotoken $$COVERALLS_TOKEN

fmt:
	[[ -z "`echo "$(GO_PACKAGES)" | xargs go fmt | tee -a /dev/stderr`" ]]

vet:
	echo $(GO_PACKAGES) | xargs go vet

lint:
	for pkg in `echo $(GO_PACKAGES)`; do golint $$pkg | grep -vE 'mock|.pb.go'; done

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
	docker build --squash -t thethingsnetwork/ttn -f Dockerfile .

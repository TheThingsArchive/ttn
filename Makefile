SHELL = bash

# Environment

GIT_BRANCH = $(or $(CI_BUILD_REF_NAME) ,`git rev-parse --abbrev-ref HEAD 2>/dev/null`)
GIT_COMMIT = $(or $(CI_BUILD_REF), `git rev-parse HEAD 2>/dev/null`)
BUILD_DATE = $(or $(CI_BUILD_DATE), `date -u +%Y-%m-%dT%H:%M:%SZ`)

# All

.PHONY: all build-deps deps dev-deps protos-clean protos mocks test cover-clean cover-deps cover coveralls fmt vet ttn ttnctl build link docs clean docker

all: deps build

# Deps

build-deps:
	@command -v govendor > /dev/null || go get "github.com/kardianos/govendor"

deps: build-deps
	govendor sync

dev-deps: deps
	@command -v protoc-gen-gofast > /dev/null || go get github.com/gogo/protobuf/protoc-gen-gofast
	@command -v mockgen > /dev/null || go get github.com/golang/mock/mockgen
	@command -v golint > /dev/null || go get github.com/golang/lint/golint
	@command -v forego > /dev/null || go get github.com/ddollar/forego

# Protobuf

PROTO_FILES = $(shell find api -name "*.proto" -and -not -name ".git")
COMPILED_PROTO_FILES = $(patsubst api%.proto, api%.pb.go, $(PROTO_FILES))
PROTOC = protoc -I/usr/local/include -I$(GOPATH)/src -I$(GOPATH)/src/github.com/TheThingsNetwork \
-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
--gofast_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:$(GOPATH)/src \
--grpc-gateway_out=:. $(GOPATH)/src/github.com/TheThingsNetwork/ttn/

protos-clean:
	rm -f $(COMPILED_PROTO_FILES)

protos: $(COMPILED_PROTO_FILES)

api/%.pb.go: api/%.proto
	$(PROTOC)$<

# Mocks

mocks:
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
	cat $(GO_COVER_FILES) | grep -v "mode: set" | sort >> $(GO_COVER_FILE)

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

splitfilename = $(subst ., ,$(subst -, ,$(subst $(RELEASE_DIR)/,,$1)))
GOOSfromfilename = $(word 2, $(call splitfilename, $1))
GOARCHfromfilename = $(word 3, $(call splitfilename, $1))
LDFLAGS = -ldflags "-w -X main.gitBranch=${GIT_BRANCH} -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}"
GOBUILD = CGO_ENABLED=0 GOOS=$(call GOOSfromfilename, $@) GOARCH=$(call GOARCHfromfilename, $@) go build -a -installsuffix cgo ${LDFLAGS} -tags "${TAGS}" -o "$@"

ttn: $(RELEASE_DIR)/ttn-$(GOOS)-$(GOARCH)$(GOEXE)

$(RELEASE_DIR)/ttn-%: $(GO_FILES)
	$(GOBUILD) ./main.go

ttnctl: $(RELEASE_DIR)/ttnctl-$(GOOS)-$(GOARCH)$(GOEXE)

$(RELEASE_DIR)/ttnctl-%: $(GO_FILES)
	$(GOBUILD) ./ttnctl/main.go

build: ttn ttnctl

GOBIN ?= $(GOPATH)/bin

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

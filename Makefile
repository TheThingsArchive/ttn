SHELL = bash

export GOOS=$(or $(word 1,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOOS`"))
export GOARCH=$(or $(word 2,$(subst -, ,${TARGET_PLATFORM})), $(shell echo "`go env GOARCH`"))
export GOEXE=$(shell echo "`GOOS=$(GOOS) GOARCH=$(GOARCH) go env GOEXE`")
export CGO_ENABLED=0

GOCMD = go
GOBUILD = $(GOCMD) build

PROTOC = protoc --gofast_out=plugins=grpc:$(GOPATH)/src/ --proto_path=$(GOPATH)/src/ $(GOPATH)/src/github.com/TheThingsNetwork/ttn

GIT_COMMIT = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date -u +%Y-%m-%dT%H:%M:%SZ`

LDFLAGS = -ldflags "-w -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}"

select_pkgs = govendor list --no-status +local
coverage_pkgs = $(select_pkgs) | grep -vE 'ttn/api|ttn/cmd|ttn/ttnctl'

RELEASE_DIR ?= release
COVER_FILE = coverage.out
TEMP_COVER_DIR ?= .cover

ttnpkg = ttn-$(GOOS)-$(GOARCH)
ttnctlpkg = ttnctl-$(GOOS)-$(GOARCH)

ttnbin = $(ttnpkg)$(GOEXE)
ttnctlbin = $(ttnctlpkg)$(GOEXE)

.PHONY: all clean build-deps deps dev-deps cover-deps vendor update-vendor proto test fmt vet cover coveralls docs build install docker package

all: clean deps build package

build-deps:
	$(GOCMD) get -u "github.com/kardianos/govendor"

deps: build-deps
	govendor sync

dev-deps: deps
	$(GOCMD) get -u -v github.com/gogo/protobuf/protoc-gen-gofast
	$(GOCMD) get -u -v github.com/golang/mock/gomock
	$(GOCMD) get -u -v github.com/golang/mock/mockgen
	$(GOCMD) get -u -v github.com/ddollar/forego

cover-deps:
	if ! $(GOCMD) get github.com/golang/tools/cmd/cover; then $(GOCMD) get golang.org/x/tools/cmd/cover; fi
	$(GOCMD) get github.com/mattn/goveralls

vendor: build-deps
	govendor add +external

update-vendor: build-deps
	govendor fetch +external

proto:
	@$(PROTOC)/api/*.proto
	@$(PROTOC)/api/protocol/protocol.proto
	@$(PROTOC)/api/protocol/**/*.proto
	@$(PROTOC)/api/gateway/gateway.proto
	@$(PROTOC)/api/router/router.proto
	@$(PROTOC)/api/broker/broker.proto
	@$(PROTOC)/api/handler/handler.proto
	@$(PROTOC)/api/networkserver/networkserver.proto
	@$(PROTOC)/api/discovery/discovery.proto
	@$(PROTOC)/api/noc/noc.proto

mocks:
	mockgen -source=./api/networkserver/networkserver.pb.go -package networkserver NetworkServerClient > api/networkserver/networkserver_mock.go
	mockgen -source=./api/discovery/client.go -package discovery NetworkServerClient > api/discovery/client_mock.go

test:
	$(select_pkgs) | xargs $(GOCMD) test

fmt:
	[[ -z "`$(select_pkgs) | xargs $(GOCMD) fmt | tee -a /dev/stderr`" ]]

vet:
	$(select_pkgs) | xargs $(GOCMD) vet

cover:
	mkdir $(TEMP_COVER_DIR)
	for pkg in $$($(coverage_pkgs)); do profile="$(TEMP_COVER_DIR)/$$(echo $$pkg | grep -oE 'ttn/.*' | sed 's/\///g').cover"; $(GOCMD) test -cover -coverprofile=$$profile $$pkg; done
	echo "mode: set" > $(COVER_FILE) && cat $(TEMP_COVER_DIR)/*.cover | grep -v mode: | sort -r | awk '{if($$1 != last) {print $$0;last=$$1}}' >> $(COVER_FILE)
	rm -r $(TEMP_COVER_DIR)

coveralls:
	$$GOPATH/bin/goveralls -coverprofile=$(COVER_FILE) -service=travis-ci -repotoken $$COVERALLS_TOKEN

clean:
	[ -d $(RELEASE_DIR) ] && rm -rf $(RELEASE_DIR) || [ ! -d $(RELEASE_DIR) ]
	([ -d $(TEMP_COVER_DIR) ] && rm -rf $(TEMP_COVER_DIR)) || [ ! -d $(TEMP_COVER_DIR) ]
	([ -f $(COVER_FILE) ] && rm $(COVER_FILE)) || [ ! -d $(COVER_FILE) ]

docs:
	cd cmd/docs && HOME='$$HOME' $(GOCMD) run generate.go > README.md
	cd ttnctl/cmd/docs && HOME='$$HOME' $(GOCMD) run generate.go > README.md

build: $(RELEASE_DIR)/$(ttnbin) $(RELEASE_DIR)/$(ttnctlbin)

install:
	$(GOCMD) install -a -installsuffix cgo ${LDFLAGS} . ./ttnctl

docker: TARGET_PLATFORM = linux-amd64
docker: clean $(RELEASE_DIR)/$(ttnbin)
	docker build -t thethingsnetwork/ttn -f Dockerfile .

package: $(RELEASE_DIR)/$(ttnpkg).zip $(RELEASE_DIR)/$(ttnpkg).tar.gz $(RELEASE_DIR)/$(ttnpkg).tar.xz $(RELEASE_DIR)/$(ttnctlpkg).zip $(RELEASE_DIR)/$(ttnctlpkg).tar.gz $(RELEASE_DIR)/$(ttnctlpkg).tar.xz

$(RELEASE_DIR)/$(ttnbin):
	$(GOBUILD) -a -installsuffix cgo ${LDFLAGS} -o $(RELEASE_DIR)/$(ttnbin) ./main.go

$(RELEASE_DIR)/$(ttnpkg).zip: $(RELEASE_DIR)/$(ttnbin)
	cd $(RELEASE_DIR) && zip -q $(ttnpkg).zip $(ttnbin)

$(RELEASE_DIR)/$(ttnpkg).tar.gz: $(RELEASE_DIR)/$(ttnbin)
	cd $(RELEASE_DIR) && tar -czf $(ttnpkg).tar.gz $(ttnbin)

$(RELEASE_DIR)/$(ttnpkg).tar.xz: $(RELEASE_DIR)/$(ttnbin)
	cd $(RELEASE_DIR) && tar -cJf $(ttnpkg).tar.xz $(ttnbin)

$(RELEASE_DIR)/$(ttnctlbin):
	$(GOBUILD) -a -installsuffix cgo ${LDFLAGS} -o $(RELEASE_DIR)/$(ttnctlbin) ./ttnctl/main.go

$(RELEASE_DIR)/$(ttnctlpkg).zip: $(RELEASE_DIR)/$(ttnctlbin)
	cd $(RELEASE_DIR) && zip -q $(ttnctlpkg).zip $(ttnctlbin)

$(RELEASE_DIR)/$(ttnctlpkg).tar.gz: $(RELEASE_DIR)/$(ttnctlbin)
	cd $(RELEASE_DIR) && tar -czf $(ttnctlpkg).tar.gz $(ttnctlbin)

$(RELEASE_DIR)/$(ttnctlpkg).tar.xz: $(RELEASE_DIR)/$(ttnctlbin)
	cd $(RELEASE_DIR) && tar -cJf $(ttnctlpkg).tar.xz $(ttnctlbin)

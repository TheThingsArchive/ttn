SHELL = bash

GOOS ?= $(shell echo "`go env GOOS`")
GOARCH ?= $(shell echo "`go env GOARCH`")
GOEXE ?= $(shell echo "`go env GOEXE`")

GOCMD = go

export CGO_ENABLED=0
GOBUILD = $(GOCMD) build

GIT_COMMIT = `git rev-parse HEAD 2>/dev/null`
BUILD_DATE = `date -u +%Y-%m-%dT%H:%M:%SZ`

LDFLAGS = -ldflags "-w -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}"

DEPS = `comm -23 <(sort <($(GOCMD) list -f '{{join .Imports "\n"}}' ./...) | uniq) <($(GOCMD) list std) | grep -v TheThingsNetwork`
TEST_DEPS = `comm -23 <(sort <($(GOCMD) list -f '{{join .TestImports "\n"}}' ./...) | uniq) <($(GOCMD) list std) | grep -v TheThingsNetwork`

select_pkgs = $(GOCMD) list ./... | grep -vE 'vendor'
coverage_pkgs = $(GOCMD) list ./... | grep -E 'core' | grep -vE 'core$$|mocks$$'

RELEASE_DIR ?= release
COVER_FILE = coverage.out
TEMP_COVER_DIR ?= .cover

ttnpkg = ttn-$(GOOS)-$(GOARCH)
ttnctlpkg = ttnctl-$(GOOS)-$(GOARCH)

ttnbin = $(ttnpkg)$(GOEXE)
ttnctlbin = $(ttnctlpkg)$(GOEXE)

.PHONY: all clean deps test-deps build-deps proto test fmt vet cover build docker package

all: clean deps build package

deps:
	$(GOCMD) get -d -v $(DEPS)

test-deps:
	$(GOCMD) get -d -v $(TEST_DEPS)

proto-deps:
	$(GOCMD) get -v github.com/gogo/protobuf/protoc-gen-gofast

proto:
	find core/protos -name '*.proto' | xargs protoc --gofast_out=plugins=grpc:./core -I=core/protos

cover-deps:
	if ! $(GOCMD) get github.com/golang/tools/cmd/cover; then $(GOCMD) get golang.org/x/tools/cmd/cover; fi
	$(GOCMD) get github.com/mattn/goveralls

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

build: $(RELEASE_DIR)/$(ttnbin) $(RELEASE_DIR)/$(ttnctlbin)

docker: GOOS = linux
docker: GOARCH = amd64
docker: clean $(RELEASE_DIR)/$(ttnbin)
	docker build -t thethingsnetwork/ttn -f Dockerfile.local .

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

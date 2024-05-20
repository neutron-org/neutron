#!/usr/bin/make -f

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app
GO_VERSION=1.22
GOLANGCI_LINT_VERSION=v1.55.2
BUILDDIR ?= $(CURDIR)/build

# for dockerized protobuf tools
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build
HTTPS_GIT := https://github.com/neutron-org/neutron.git

GO_SYSTEM_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1-2)
REQUIRE_GO_VERSION = $(GO_VERSION)

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

build_tags_test_binary = $(build_tags)
build_tags_test_binary += skip_ccv_msg_filter

whitespace :=
empty = $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(empty),$(comma),$(build_tags))
build_tags_test_binary_comma_sep := $(subst $(empty),$(comma),$(build_tags_test_binary))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=neutron \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=neutrond \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)' -trimpath
BUILD_FLAGS_TEST_BINARY := -tags "$(build_tags_test_binary_comma_sep)" -ldflags '$(ldflags)' -trimpath

# The below include contains the tools and runsim targets.
include contrib/devtools/Makefile

check_version:
ifneq ($(GO_SYSTEM_VERSION), $(REQUIRE_GO_VERSION))
	@echo "ERROR: Go version ${REQUIRE_GO_VERSION} is required for $(VERSION) of Neutron."
	exit 1
endif

all: install lint test

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): check_version go.sum $(BUILDDIR)/
ifeq ($(OS),Windows_NT)
	exit 1
else
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./cmd/neutrond
endif

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

build-static-linux-amd64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name neutronbuilder || true
	$(DOCKER) buildx use neutronbuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TAGS=$(build_tags_comma_sep),muslc \
		--platform linux/amd64 \
		-t neutron-amd64 \
		--load \
		-f Dockerfile.builder .
	$(DOCKER) rm -f neutronbinary || true
	$(DOCKER) create -ti --name neutronbinary neutron-amd64
	$(DOCKER) cp neutronbinary:/bin/neutrond $(BUILDDIR)/neutrond-linux-amd64
	$(DOCKER) rm -f neutronbinary

build-slinky-e2e-docker-image: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name neutronbuilder || true
	$(DOCKER) buildx use neutronbuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TAGS=$(build_tags_comma_sep),skip_ccv_msg_filter \
		--build-arg RUNNER_IMAGE="alpine:3.18" \
		--platform linux/amd64 \
		-t neutron-e2e \
		--load \
		-f Dockerfile.builder .

slinky-e2e-test: build-slinky-e2e-docker-image
	cd ./tests/slinky && go mod tidy && go test -v -race -timeout 20m ./...

install-test-binary: check_version go.sum
	go install -mod=readonly $(BUILD_FLAGS_TEST_BINARY) ./cmd/neutrond

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/neutrond -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf snapcraft-local.yaml $(BUILDDIR)/

distclean: clean
	rm -rf vendor/

########################################
### Testing


test: test-unit
	@rm -rf ./.testchains

test-all: check test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

test-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

benchmark:
	@go test -mod=readonly -bench=. ./...

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppImportExport

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 10 TestFullAppSimulation

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*_test.go" | xargs gofmt -d -s

format:
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name "*.pb.go" -not -name "*.pb.gw.go" -not -name "*.pulsar.go" -not -path "./crypto/keys/secp256k1/*" | xargs gofumpt -w -l
	golangci-lint run --fix
	goimports -w -local github.com/neutron-org .

.PHONY: format

###############################################################################
###                                Protobuf                                 ###
###############################################################################
protoVer=0.11.6
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=docker run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)
proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-swagger-gen:
	@$(DOCKER) build proto/ -t swagger-gen
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace swagger-gen ./scripts/protoc-swagger-gen.sh

PROTO_FORMATTER_IMAGE=bufbuild/buf:1.28.1

proto-format:
	@echo "Formatting Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace \
	--workdir /workspace $(PROTO_FORMATTER_IMAGE) \
	 format proto -w


.PHONY: all install install-debug \
	go-mod-cache draw-deps clean build format \
	test test-all test-build test-cover test-unit test-race \
	test-sim-import-export \

init: kill-dev install-test-binary
	@echo "Building gaiad binary..."
	@cd ./../gaia/ && make install
	@echo "Initializing both blockchains..."
	./network/init-and-start-both.sh
	@echo "Initializing relayer..."
	./network/hermes/restore-keys.sh
	./network/hermes/create-conn.sh

start: kill-dev install-test-binary
	@echo "Starting up neutrond alone..."
	export BINARY=neutrond CHAINID=test-1 P2PPORT=26656 RPCPORT=26657 RESTPORT=1317 ROSETTA=8080 GRPCPORT=8090 GRPCWEB=8091 STAKEDENOM=untrn && \
	./network/init.sh && ./network/init-neutrond.sh && ./network/start.sh

start-rly:
	./network/hermes/start.sh

kill-dev:
	@echo "Killing neutrond and removing previous data"
	-@rm -rf ./data
	-@killall neutrond 2>/dev/null
	-@killall gaiad 2>/dev/null

build-docker-image:
	# please keep the image name consistent with https://github.com/neutron-org/neutron-integration-tests/blob/main/setup/docker-compose.yml
	@docker buildx build --load --build-context app=. -t neutron-node --build-arg BINARY=neutrond .

start-docker-container:
	# please keep the ports consistent with https://github.com/neutron-org/neutron-integration-tests/blob/main/setup/docker-compose.yml
	@docker run --rm --name neutron -d -p 1317:1317 -p 26657:26657 -p 26656:26656 -p 16657:16657 -p 8090:9090 -e RUN_BACKGROUND=0 neutron-node

stop-docker-container:
	@docker stop neutron

mocks:
	@echo "Regenerate mocks..."
	@go generate ./...

format-all: format proto-format

check-proto-format:
	@echo "Formatting Protobuf files"
		$(DOCKER) run --rm -v $(CURDIR):/workspace \
		--workdir /workspace $(PROTO_FORMATTER_IMAGE) \
		format proto -d --exit-code

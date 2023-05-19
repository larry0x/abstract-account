#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build
LEDGER_ENABLED ?= true

# ********** Golang configs **********

export GO111MODULE = on

# the Go version to use in reproducible build
GO_VERSION := 1.20

# currently installed Go version
GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

# minimum supported Go version
GO_MINIMUM_MAJOR_VERSION = $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f2 | cut -d'.' -f1)
GO_MINIMUM_MINOR_VERSION = $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f2 | cut -d'.' -f2)

GO_VERSION_ERR_MSG = ❌ ERROR: Go version $(GO_MINIMUM_MAJOR_VERSION).$(GO_MINIMUM_MINOR_VERSION)+ is required

# ********** process build tags **********

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

ifeq (cleveldb,$(findstring cleveldb,$(MARS_BUILD_OPTIONS)))
	build_tags += gcc cleveldb
else ifeq (rocksdb,$(findstring rocksdb,$(MARS_BUILD_OPTIONS)))
	build_tags += gcc rocksdb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# ********** process linker flags **********

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=abstract-account \
          -X github.com/cosmos/cosmos-sdk/version.AppName=aad \
          -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
          -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
          -X github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)

ifeq (cleveldb,$(findstring cleveldb,$(MARS_BUILD_OPTIONS)))
	ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
else ifeq (rocksdb,$(findstring rocksdb,$(MARS_BUILD_OPTIONS)))
	ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
ifeq (,$(findstring nostrip,$(MARS_BUILD_OPTIONS)))
	ldflags += -w -s
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags '$(build_tags)' -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(MARS_BUILD_OPTIONS)))
	BUILD_FLAGS += -trimpath
endif

all: proto-gen lint test install

################################################################################
###                                  Build                                   ###
################################################################################

install: enforce-go-version
	@echo "🤖 Installing aad..."
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/aad
	@echo "✅ Completed installation!"

build: enforce-go-version
	@echo "🤖 Building aad..."
	go build $(BUILD_FLAGS) -o $(BUILDDIR)/ ./cmd/aad
	@echo "✅ Completed build!"

# For use when publishing releases
#
# Copied from Osmosis:
# https://github.com/osmosis-labs/osmosis/blob/v14.0.1/Makefile#L111
build-reproducible: build-reproducible-amd64 build-reproducible-arm64

build-reproducible-amd64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name aabuilder || true
	$(DOCKER) buildx use aabuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.17 \
		--platform linux/amd64 \
		-t aa:local-amd64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f aabinary || true
	$(DOCKER) create -ti --name aabinary aa:local-amd64
	$(DOCKER) cp aabinary:/bin/aad $(BUILDDIR)/aad-linux-amd64
	$(DOCKER) rm -f aabinary

build-reproducible-arm64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name aabuilder || true
	$(DOCKER) buildx use aabuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg RUNNER_IMAGE=alpine:3.17 \
		--platform linux/arm64 \
		-t aa:local-arm64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f aabinary || true
	$(DOCKER) create -ti --name aabinary aa:local-arm64
	$(DOCKER) cp aabinary:/bin/aad $(BUILDDIR)/aad-linux-arm64
	$(DOCKER) rm -f aabinary

# Make sure that Go version is 1.19+
#
# From Osmosis discord:
# https://discord.com/channels/798583171548840026/837144686387920936/1049449765240315925
#
# > Valardragon - 12/05/2022 10:18 PM
# > It was just pointed out from `@jhernandez | stargaze.zone`, that the choice
#   of golang version between go 1.18 and go 1.19 is consensus critical.
#   With insufficient info, this preliminarily seems due to go 1.19 changing the
#   memory model format, and something state-affecting in cosmwasm getting altered.
#   https://github.com/persistenceOne/incident-reports/blob/main/06-nov-2022_V4_upgrade_halt.md
enforce-go-version:
	@echo "🤖 Go version: $(GO_MAJOR_VERSION).$(GO_MINOR_VERSION)"
	@if [ $(GO_MAJOR_VERSION) -gt $(GO_MINIMUM_MAJOR_VERSION) ]; then \
		exit 0; \
	elif [ $(GO_MAJOR_VERSION) -lt $(GO_MINIMUM_MAJOR_VERSION) ]; then \
		echo '$(GO_VERSION_ERR_MSG)'; \
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -lt $(GO_MINIMUM_MINOR_VERSION) ]; then \
		echo '$(GO_VERSION_ERR_MSG)'; \
		exit 1; \
	fi

################################################################################
###                                  Tests                                   ###
################################################################################

test:
	@echo "🤖 Running tests..."
	go test -mod=readonly ./x/...
	@echo "✅ Completed tests!"

################################################################################
###                                 Protobuf                                 ###
################################################################################

# We use osmolabs' docker image instead of tendermintdev/sdk-proto-gen.
# The only difference is that the Osmosis version uses Go 1.19 while the
# tendermintdev one uses 1.18.
protoVer=v0.9
protoImageName=osmolabs/osmo-proto-gen:$(protoVer)
containerProtoGenGo=aa-proto-gen-go-$(protoVer)
containerProtoGenSwagger=aa-proto-gen-swagger-$(protoVer)

proto-gen: proto-go-gen proto-swagger-gen

proto-go-gen:
	@echo "🤖 Generating Go code from protobuf..."
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenGo}$$"; then docker start -a $(containerProtoGenGo); else docker run --name $(containerProtoGenGo) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi
	@echo "✅ Completed Go code generation!"

proto-swagger-gen:
	@echo "🤖 Generating Swagger code from protobuf..."
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenSwagger}$$"; then docker start -a $(containerProtoGenSwagger); else docker run --name $(containerProtoGenSwagger) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protoc-swagger-gen.sh; fi
	@echo "✅ Completed Swagger code generation!"

################################################################################
###                                  Docker                                  ###
################################################################################

RUNNER_BASE_IMAGE_DISTROLESS := gcr.io/distroless/static-debian11
RUNNER_BASE_IMAGE_ALPINE := alpine:3.17
RUNNER_BASE_IMAGE_NONROOT := gcr.io/distroless/static-debian11:nonroot

docker-build:
	@DOCKER_BUILDKIT=1 $(DOCKER) build \
		-t aa:local \
		-t aa:local-distroless \
		--build-arg GO_VERSION=$(GO_MAJOR_VERSION).$(GO_MINOR_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_DISTROLESS) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f Dockerfile .

docker-build-alpine:
	@DOCKER_BUILDKIT=1 $(DOCKER) build \
		-t aa:local-alpine \
		--build-arg GO_VERSION=$(GO_MAJOR_VERSION).$(GO_MINOR_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_ALPINE) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f Dockerfile .

docker-build-nonroot:
	@DOCKER_BUILDKIT=1 $(DOCKER) build \
		-t aa:local-nonroot \
		--build-arg GO_VERSION=$(GO_MAJOR_VERSION).$(GO_MINOR_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_NONROOT) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f Dockerfile .

docker-build-distroless: docker-build

################################################################################
###                                 Linting                                  ###
################################################################################

golangci_lint_cmd=github.com/golangci/golangci-lint/cmd/golangci-lint

lint:
	@echo "🤖 Running linter..."
	go run $(golangci_lint_cmd) run --timeout=10m
	@echo "✅ Completed linting!"

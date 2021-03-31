# Init variables
VERSION ?= $(shell git describe --tags --always)
COMMIT ?= $(shell git rev-parse HEAD)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

DEP_IMAGE_NAME ?= confluentinc/cp-schema-registry:6.1.1
DEP_CONTAINER_NAME ?= registry-test

LDFLAGS = "-w -X main.Version=$(VERSION) -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

GITHUB_ORG_NAME = bjornm82
GITHUB_PROJECT = schema-registry
PROJECT_DIR ?= src/github.com/${GITHUB_ORG_NAME}/${GITHUB_PROJECT}

APP ?= schema-registry
OS ?= linux
ARCH ?= amd64

.PHONY: build
build:
	@docker build . -f Dockerfile.build \
		--build-arg PROJECT_DIR=${PROJECT_DIR}

.PHONY: test-unit
test-unit:
	go test -race -v ./...

.PHONY: test-integration
test-integration:
	PROJECT_DIR=src/github.com/bjornm82 CONFLUENT_VERSION=6.1.0 docker-compose up --build --exit-code-from client && \
	docker-compose down

# .PHONY: benchmark
# benchmark:
# 	go test -bench=. -benchmem -cpuprofile profile.out -memprofile memory.out ./internal/

.PHONY: godoc
# Godoc command, in order to check the package go to:
# http://localhost:6060/pkg/github.com/rtlnl/di-pixel/
# after running the following command.
godoc:
	godoc -http=:6060

.PHONY: all
all:
	$(MAKE) build
	docker-compose up --build
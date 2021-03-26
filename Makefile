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

.PHONY: test-all test-db-run test-integration test-db-kill
test-all: test-unit test-db-run test-integration test-db-kill

.PHONY: test-unit
test-unit:
	go test -race -v ./...

.PHONY: test-db-run
test-db-run:
	docker run --rm -d --name ${DEP_CONTAINER_NAME} -e SCHEMA_REGISTRY_HOST_NAME=schema-registry -e SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS='broker:29092' -e SCHEMA_REGISTRY_LISTENERS=http://0.0.0.0:8081 -p 8086:8081 ${DEP_IMAGE_NAME}

.PHONY: test-integration
test-integration:
	go test -tags=integration

.PHONY: test-db-kill
test-db-kill:
	docker kill ${DEP_CONTAINER_NAME}

# .PHONY: benchmark
# benchmark:
# 	go test -bench=. -benchmem -cpuprofile profile.out -memprofile memory.out ./internal/

.PHONY: test
test:
	go test -race -v ./...

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
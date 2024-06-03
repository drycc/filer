SHORT_NAME ?= filer

include versioning.mk
DRYCC_REGISTRY ?= ${DEV_REGISTRY}

# container development environment variables
REPO_PATH := github.com/drycc/${SHORT_NAME}
DEV_ENV_IMAGE := ${DEV_REGISTRY}/drycc/go-dev
DEV_ENV_WORK_DIR := /root/go/src/${REPO_PATH}
DEV_ENV_PREFIX := podman run --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR}
DEV_ENV_CMD := ${DEV_ENV_PREFIX} ${DEV_ENV_IMAGE}
PLATFORM ?= linux/amd64,linux/arm64

# Common flags passed into Go's linker.
LDFLAGS := "-s -w -X main.version=${VERSION}"

bootstrap:
	${DEV_ENV_CMD} go mod vendor

test: test-style test-unit

test-style:
	${DEV_ENV_CMD} lint

test-unit:
	${DEV_ENV_CMD} go test --race ./...

test-cover:
	${DEV_ENV_CMD} test-cover.sh

podman-build:
	podman build -t ${IMAGE} --build-arg LDFLAGS=${LDFLAGS} --build-arg CODENAME=${CODENAME} .
	podman tag ${IMAGE} ${MUTABLE_IMAGE}

.PHONY: bootstrap podman-build test test-style test-unit test-cover
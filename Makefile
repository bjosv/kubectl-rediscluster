BINARY := kubectl-rediscluster
VERSION := v0.0.1
GIT := $(shell git rev-parse HEAD || echo "")
GOVERSION := $(shell go version || echo "")


all: build test

build:
	GO111MODULE="on" CGO_ENABLED=0 \
	go build -o $(BINARY) \
	-ldflags='-X main.version=$(VERSION) -X main.git=$(GIT) -X "main.goversion=$(GOVERSION)"' \
	cmd/$(BINARY).go \

test:
	go test -v -short -race -timeout 30s ./...

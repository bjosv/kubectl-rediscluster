BINARY := kubectl-rediscluster
VERSION := v0.0.1
COMMIT := $(shell git rev-parse HEAD || echo "")

all: build test

build:
	GO111MODULE="on" CGO_ENABLED=0 \
	go build -o $(BINARY) \
	-ldflags='-X main.version=$(VERSION) -X main.commit=$(COMMIT)' \
	cmd/$(BINARY).go \

test:
	go test -v -short ./...
# go test -v -short -race -timeout 30s ./...

clean:
	go clean -r

release:
	goreleaser --snapshot --skip-publish --rm-dist

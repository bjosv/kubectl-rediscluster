
version := v0.0.1
git := $(shell git rev-parse HEAD || echo "")
goversion := $(shell go version || echo "")

all:
	GO111MODULE="on" go build \
	-o kubectl-rediscluster \
	-ldflags='-X main.version=$(version) -X main.git=$(git) -X "main.goversion=$(goversion)"' \
	cmd/kubectl-rediscluster/main.go \

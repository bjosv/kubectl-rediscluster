# kubectl-rediscluster

## Build

```
GO111MODULE="on" go build

# or to bake in a plugin version:

GO111MODULE="on" go build -ldflags="-X main.version=v1.0.0"
```

## Setup

Use something like:

```
ln -s $PWD/kubectl-rediscluster ~/bin/
alias kc="kubectl rediscluster"

# Verify with
kubectl plugin list
```

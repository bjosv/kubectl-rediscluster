# kubectl-rediscluster

## Build

```bash
make
```

## Setup

A kubectl plugin binary needs to accessable via $PATH

When ~/bin is in your path, use something like:

```bash
ln -s $PWD/kubectl-rediscluster ~/bin/

# Verify with
kubectl plugin list

# Avoid long names by creating a alias like
alias kc="kubectl rediscluster"
```

## Commands

### Get slots information

Get the slot distribution of a Redis Cluster service named `cluster-redis-cluster`

```bash
kubectl rediscluster slots cluster-redis-cluster
```

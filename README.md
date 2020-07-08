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
If no service name is given it will be guessed.

```bash
kubectl rediscluster slots cluster-redis-cluster

START  END    MASTER           REPLICA          PODNAME                     NODE          REMARKS
0      5461   10.244.2.2:6379                   rediscluster-cluster-dqrzl  kind-worker
.      .                       10.244.3.3:6379  rediscluster-cluster-lvkmz  kind-worker2
5462   10923  10.244.1.3:6379                   rediscluster-cluster-7tpnv  kind-worker3
.      .                       10.244.1.4:6379  rediscluster-cluster-kgtrm  kind-worker3
10924  16383  10.244.2.3:6379                   rediscluster-cluster-8gsfm  kind-worker
.      .                       10.244.3.2:6379  rediscluster-cluster-v7dcl  kind-worker2
```

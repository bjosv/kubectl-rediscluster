# kubectl-rediscluster

A kubectl plugin for inspecting your Redis Cluster. The plugin will collect information from the K8s API and by connecting to the Redis instances within the pods.

## Installation

Download the binary, or fetch it via

`GO111MODULE=on go get github.com/bjosv/kubectl-rediscluster@latest`

A kubectl plugin binary needs to accessable via $PATH, so make sure its in the path and verify the installation by running: `kubectl plugin list`

### Options

```bash
# Avoid long names by creating an alias like
alias kc="kubectl rediscluster"

# ..or rename the base command like:
ln -s ~/bin/kubectl-rediscluster ~/bin/kubectl-rc
kubectl rc slots
```

## Commands

### Get slots information

Get the slot distribution of a Redis Cluster service named `cluster-redis-cluster`

Example:

```bash
> kubectl rediscluster slots cluster-redis-cluster

START  END    MASTER           REPLICA          PODNAME                     HOST          REMARKS
0      5461   10.244.2.2:6379                   rediscluster-cluster-dqrzl  kind-worker
.      .                       10.244.3.3:6379  rediscluster-cluster-lvkmz  kind-worker2
5462   10923  10.244.1.3:6379                   rediscluster-cluster-7tpnv  kind-worker3  *Same host*
.      .                       10.244.1.4:6379  rediscluster-cluster-kgtrm  kind-worker3  *Same host*
10924  16383  10.244.3.2:6379                   rediscluster-cluster-v7dcl  kind-worker2  *Replica missing*
```

### Options

If the service name is skipped the plugin will try to guess which service that handles the Redis cluster.

Example:

```bash
> kubectl rediscluster slots
```

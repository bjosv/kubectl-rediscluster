# kubectl-rediscluster

A kubectl plugin for inspecting your Redis Clusters deployed in Kubernetes. The plugin will collect information from the K8s API, and by connecting to Redis instances within the pods.

The plugin expects a K8s Service to be able to find out which Pods that runs Redis in cluster mode.

## Installation

Download the binary, or fetch it via:

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

Get the slot distribution of a Redis Cluster, with additional information about pods and placement on the K8s hosts. Some analyzis regarding fault tolerance is shown in the remarks column.

`kubectl rediscluster slots <SERVICE NAME>`

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

#### Omit service name

Let the plugin guess which service that provides the Redis Cluster by omitting the service name.
A service that provides the port 6379 will be selected.

Example:

```bash
> kubectl rediscluster slots
```

#### Set namespace

All the usual kubeconfig options are available, like using specific namespace. See help (-h)

Example:

```bash
> kubectl rediscluster slots -n mynamespace
```

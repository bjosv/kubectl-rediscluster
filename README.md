# kubectl-rediscluster

A kubectl plugin for inspecting your Redis Clusters deployed in Kubernetes. The plugin will collect information from the K8s API, and by querying the Redis instances within the pods.

The plugin expects a K8s Service to be able to find out which Pods that runs Redis in cluster mode.

## Installation

Download the binary, or fetch it via:

`GO111MODULE=on go get github.com/bjosv/kubectl-rediscluster@latest`

A kubectl plugin binary needs to be accessible via $PATH, so make sure its in the path and verify the installation by running: `kubectl plugin list`

#### Options

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

START  END    ROLE    IP               PODNAME                     HOST          REMARKS
0      5461   Master  10.244.3.3:6379  rediscluster-cluster-lvkmz  kind-worker2
.      .      Repl.   10.244.2.2:6379  rediscluster-cluster-dqrzl  kind-worker
5462   10923  Master  10.244.1.3:6379  rediscluster-cluster-7tpnv  kind-worker3  *Same host*
.      .      Repl.   10.244.1.4:6379  rediscluster-cluster-kgtrm  kind-worker3  *Same host*
10924  16383  Master  10.244.3.2:6379  rediscluster-cluster-v7dcl  kind-worker2  *Replica missing*
```

### Get nodes information

Get information about the Redis Cluster instances.

`kubectl rediscluster nodes <SERVICE NAME>`

Example:

```bash
> kubectl rediscluster nodes cluster-redis-cluster
                                                                         SLOT    CLUSTER
HOST          PODNAME                     IP          ROLE  KEYS  SLOTS  RANGES  STATE    REMARKS
kind-worker   rediscluster-cluster-pf24w  10.244.1.3  ?     0     5462   1       ok
kind-worker   rediscluster-cluster-vkttp  10.244.1.4  ?     0     5462   1       ok
kind-worker2  rediscluster-cluster-rfvsr  10.244.2.2  ?     0     5460   1       ok
kind-worker2  rediscluster-cluster-kpxf2  10.244.2.3  ?     0     5460   1       ok
kind-worker3  rediscluster-cluster-j8kd4  10.244.3.2  ?     0     5462   1       ok
kind-worker3  rediscluster-cluster-4rxkj  10.244.3.3  ?     0     5462   1       ok
```

### Options

#### Omit service name

Let the plugin guess which service that provides the Redis Cluster by omitting the service name.
A service that provides the port 6379 will be selected.

Example:

```bash
> kubectl rediscluster slots
Using service name: cluster-redis-cluster
...
```

#### Set namespace

All the usual kubeconfig options are available, like using specific namespace. See help (-h)

Example:

```bash
> kubectl rediscluster slots -n mynamespace
```

#### Verbose logging

```bash
> kubectl rediscluster slots -v
```

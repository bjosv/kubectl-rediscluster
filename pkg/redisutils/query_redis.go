package redisutils

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/bjosv/kubectl-rediscluster/pkg/portforwarder"
	"github.com/go-redis/redis/v8"
)

const RedisPort = 6379

type ClusterInfo map[string]string
type ClusterNodes map[string][]string
type ClusterSlots []redis.ClusterSlot

// Sorter functions: Sort by Start slot
type BySlot ClusterSlots

func (s BySlot) Len() int           { return len(s) }
func (s BySlot) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s BySlot) Less(i, j int) bool { return s[i].Start < s[j].Start }

func QueryRedis(pfwd *portforwarder.PortForwarder, namespace string, podName string, podPort int) (ClusterInfo, ClusterNodes, ClusterSlots, error) {

	localPort, err := portforwarder.GetAvailableLocalPort()
	if err != nil {
		return nil, nil, nil, err
	}

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		pfwd.ForwardPort(namespace, podName, localPort, podPort, stopCh, readyCh)
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	// Wait for portforwaring to be ready
	select {
	case <-readyCh:
		break
	}

	// Connect to Redis instance in pod (using portforwarding)
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", localPort),
	})
	var ctx = context.Background()

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}

	slots, err := rdb.ClusterSlots(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}
	sort.Sort(BySlot(slots))

	cInfo, err := rdb.ClusterInfo(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}

	// Structure the cluster info data
	info := make(map[string]string)
	for _, line := range strings.Split(cInfo, "\n") {
		keyVals := strings.Split(line, ":")

		if len(keyVals) == 2 {
			info[keyVals[0]] = keyVals[1]
		}
	}

	dbSize, err := rdb.DBSize(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}
	info["keys"] = fmt.Sprintf("%d", dbSize)

	cNodes, err := rdb.ClusterNodes(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}

	// Structure the cluster info data
	nodes := make(map[string][]string)
	for _, line := range strings.Split(cNodes, "\n") {
		keyVals := strings.Split(line, " ")

		if len(keyVals) > 6 {
			addr := strings.Split(keyVals[1], ":")
			if len(addr) > 1 {
				ip := addr[0]
				nodes[ip] = keyVals
			}
		}
	}

	// Stop and wait for portforwarder goroutine to exit
	close(stopCh)
	wg.Wait()

	return info, nodes, slots, nil
}

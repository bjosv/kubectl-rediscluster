package redisutils

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bjosv/kubectl-rediscluster/pkg/portforwarder"
	"github.com/go-redis/redis/v8"
)

const RedisPort = 6379
const Timeout = 2

type RedisInfo map[string]string
type ClusterSlots []redis.ClusterSlot

// Sorter functions: Sort by Start slot
type BySlot ClusterSlots

func (s BySlot) Len() int           { return len(s) }
func (s BySlot) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s BySlot) Less(i, j int) bool { return s[i].Start < s[j].Start }

func QueryRedis(pfwd *portforwarder.PortForwarder, namespace string, podName string, podPort int) (RedisInfo, ClusterNodes, ClusterSlots, error) {
	localPort, err := portforwarder.GetAvailableLocalPort()
	if err != nil {
		return nil, nil, nil, err
	}

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})
	errorCh := make(chan error)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		err = pfwd.ForwardPort(namespace, podName, localPort, podPort, stopCh, readyCh)
		if err != nil {
			errorCh <- err
		}
		wg.Done()
	}()

	// Wait for portforwaring to be ready
	select {
	case <-readyCh:
		break
	case err := <-errorCh:
		close(stopCh)
		return nil, nil, nil, err
	case <-time.After(Timeout * time.Second):
		close(stopCh)
		return nil, nil, nil, fmt.Errorf("could not setup a portforward to %s/%s:%d", namespace, podName, podPort)
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

	rInfo, err := rdb.Info(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}

	dbSize, err := rdb.DBSize(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}

	cNodes, err := rdb.ClusterNodes(ctx).Result()
	if err != nil {
		return nil, nil, nil, err
	}
	// Done with the portforwarder
	close(stopCh)

	// Parse query responses
	info := make(map[string]string)
	for _, line := range strings.Split(cInfo, "\r\n") {
		keyVals := strings.Split(line, ":")

		if len(keyVals) > 1 {
			info[keyVals[0]] = keyVals[1]
		}
	}
	for _, line := range strings.Split(rInfo, "\r\n") {
		keyVals := strings.Split(line, ":")

		if len(keyVals) > 1 {
			info[keyVals[0]] = keyVals[1]
		}
	}
	info["keys"] = fmt.Sprintf("%d", dbSize)

	// Parse cluster nodes data
	nodes := NewClusterNodes(cNodes)

	// Wait for portforwarder goroutine to exit
	wg.Wait()

	return info, nodes, slots, nil
}

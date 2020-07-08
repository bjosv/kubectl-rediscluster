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

// Sorter type
type BySlot []redis.ClusterSlot

type ClusterInfo map[string]string

// Sorter functions
func (s BySlot) Len() int           { return len(s) }
func (s BySlot) Less(i, j int) bool { return s[i].Start < s[j].Start }
func (s BySlot) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func QueryRedis(pfwd *portforwarder.PortForwarder, namespace string, podName string, podPort int) ([]redis.ClusterSlot, ClusterInfo, error) {

	localPort, err := portforwarder.GetAvailableLocalPort()
	if err != nil {
		return nil, nil, err
	}
	//fmt.Printf("Get CLUSTER SLOTS from %s/%s using localport: %d\n", namespace, podName, localPort)

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
		return nil, nil, err
	}

	slots, err := rdb.ClusterSlots(ctx).Result()
	if err != nil {
		return nil, nil, err
	}
	sort.Sort(BySlot(slots))

	cInfo, err := rdb.ClusterInfo(ctx).Result()
	if err != nil {
		return nil, nil, err
	}

	info := make(map[string]string)
	for _, line := range strings.Split(cInfo, "\n") {
		keyVals := strings.Split(line, ":")

		if len(keyVals) == 2 {
			info[keyVals[0]] = keyVals[1]
		}
	}

	dbSize, err := rdb.DBSize(ctx).Result()
	if err != nil {
		return nil, nil, err
	}
	info["keys"] = fmt.Sprintf("%d", dbSize)

	// Stop and wait for portforwarder goroutine to exit
	close(stopCh)
	wg.Wait()

	return slots, info, nil
}

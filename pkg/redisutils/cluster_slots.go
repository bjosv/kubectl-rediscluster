package redisutils

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/bjosv/kubectl-rediscluster/pkg/portforwarder"
	"github.com/go-redis/redis/v8"
)

// Sorter type
type BySlot []redis.ClusterSlot

// Sorter functions
func (s BySlot) Len() int           { return len(s) }
func (s BySlot) Less(i, j int) bool { return s[i].Start < s[j].Start }
func (s BySlot) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func GetClusterSlots(pfwd *portforwarder.PortForwarder, namespace string, podName string, podPort int) ([]redis.ClusterSlot, error) {

	localPort, err := portforwarder.GetAvailableLocalPort()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Get CLUSTER SLOTS from %s/%s using localport: %d\n", namespace, podName, localPort)

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
		return nil, err
	}

	slots, err := rdb.ClusterSlots(ctx).Result()
	if err != nil {
		return nil, err
	}
	// Done
	close(stopCh)

	sort.Sort(BySlot(slots))

	wg.Wait()
	//println("Done...")

	return slots, nil
}

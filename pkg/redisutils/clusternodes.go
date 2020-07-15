package redisutils

import (
	"strings"
)

// Structure of CLUSTER NODES result:
// id, ip:port@port, flags(self/master..), master-id, ping, pong, config-epoch, linkstate, slot

// ClusterNodes is a type that holds the result from one query
type ClusterNodes map[string][]string

const MinElements = 6

func NewClusterNodes(cliData string) ClusterNodes {
	nodes := make(map[string][]string)
	for _, line := range strings.Split(cliData, "\n") {
		keyVals := strings.Split(line, " ")

		if len(keyVals) > MinElements {
			addr := strings.Split(keyVals[1], ":")
			if len(addr) > 1 {
				ip := addr[0]
				nodes[ip] = keyVals
			}
		}
	}
	return nodes
}

// GetFlagsSelf returns the status field for self
func (n *ClusterNodes) GetFlagsSelf() string {
	for _, status := range *n {
		if strings.Contains(status[2], "myself") {
			flags := strings.Replace(status[2], "myself,", "", -1)
			return strings.Replace(flags, "slave", "repl.", -1)
		}
	}
	return ""
}

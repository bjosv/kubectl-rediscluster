package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/bjosv/kubectl-rediscluster/pkg/k8s"
	"github.com/bjosv/kubectl-rediscluster/pkg/portforwarder"
	"github.com/bjosv/kubectl-rediscluster/pkg/redisutils"
	"github.com/go-redis/redis/v8"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type slotsCmd struct {
	configFlags *genericclioptions.ConfigFlags
	streams     *genericclioptions.IOStreams
	args        []string

	k8sInfo    *k8s.ClusterInfo
	redisInfo  map[string]redisutils.ClusterInfo
	redisSlots map[string]redisutils.ClusterSlots
	verbose    bool
}

// NewSlotsCmd initialize and creates a Cobra command
func NewSlotsCmd(streams genericclioptions.IOStreams) *cobra.Command {
	c := &slotsCmd{
		configFlags: genericclioptions.NewConfigFlags(true),
		streams:     &streams,
		k8sInfo:     k8s.NewClusterInfo(),
		redisInfo:   make(map[string]redisutils.ClusterInfo),
		redisSlots:  make(map[string]redisutils.ClusterSlots),
	}

	cmd := &cobra.Command{
		Use:   "slots [service-name] [flags]",
		Short: "Show slots distribution of a Redis Cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.Complete(cmd, args); err != nil {
				return err
			}
			if err := c.Validate(); err != nil {
				return err
			}
			cmd.SilenceUsage = true // No usage if Run() fails, like missing service
			if err := c.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	// Add kubectl config flags to this command
	c.configFlags.AddFlags(cmd.Flags())

	cmd.Flags().BoolVarP(&c.verbose, "verbose", "v", false, "Show verbose logs")
	return cmd
}

// Complete sets all information required for the command
func (c *slotsCmd) Complete(cmd *cobra.Command, args []string) error {
	c.args = args

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (c *slotsCmd) Validate() error {
	if len(c.args) > 1 {
		return fmt.Errorf("maximum 1 service name can be given, got %d", len(c.args))
	}

	return nil
}

// Run the command
func (c *slotsCmd) Run() error {
	namespace, err := k8s.CurrentNamespace(c.configFlags)
	if err != nil {
		return err
	}

	restConfig, err := c.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	serviceName := ""
	if len(c.args) > 0 {
		serviceName = c.args[0]
	} else {
		serviceName, err = k8s.FindServiceUsingPort(restConfig, namespace, redisutils.RedisPort)
		if err != nil {
			return fmt.Errorf("%s\nPlease provide a service name", err)
		}
		fmt.Fprintf(c.streams.Out, "Using service name: %s\n", serviceName)
	}

	// Get pod info
	err = getK8sInfo(restConfig, serviceName, namespace, c.k8sInfo)
	if err != nil {
		return err
	}

	var pfwd *portforwarder.PortForwarder
	if c.verbose {
		pfwd = portforwarder.New(restConfig, c.streams.Out, c.streams.ErrOut)
	} else {
		pfwd = portforwarder.New(restConfig, nil, nil)

		// Silence K8s errors, like connection refuse
		logKubeError := func(err error) {}
		runtime.ErrorHandlers = []func(error){logKubeError}
	}

	// Query all pods/redis instances
	ch := make(chan QueryRedisResult)
	for _, pod := range c.k8sInfo.Pods {
		go func(podName string, podPort int, ch chan QueryRedisResult) {
			clusterInfo, _, clusterSlots, err := redisutils.QueryRedis(pfwd, namespace, podName, podPort)
			if err != nil {
				fmt.Printf("Failed to get Redis Cluster slot information for pod=%s: %v\n",
					podName, err)
			}
			ch <- QueryRedisResult{
				PodName: podName,
				Info:    clusterInfo,
				Slots:   clusterSlots,
			}
		}(pod.Name, redisutils.RedisPort, ch)
	}

	// Collect results from all pods/redis instances
	for range c.k8sInfo.Pods {
		queryResult := <-ch

		if queryResult.Info != nil {
			c.redisInfo[queryResult.PodName] = queryResult.Info
		}

		if queryResult.Slots != nil {
			c.redisSlots[queryResult.PodName] = queryResult.Slots
		}
	}

	//	Display result
	c.outputResult()

	return nil
}

func analyzeSlotsInfo(slots redis.ClusterSlot, info *k8s.ClusterInfo) string {
	result := ""
	// Check redundancy
	if len(slots.Nodes) == 1 {
		result += "*replica missing*"
	}
	// Check distribution on K8s workers
	if len(slots.Nodes) > 1 {
		host := ""
		for i, node := range slots.Nodes {
			h := info.GetPodInfo(node.Addr).Host
			if i == 0 {
				host = h
			} else if h != host {
				host = ""
				break // Found difference, skip rest
			}
		}
		if host != "" {
			result += "*same host*"
		}
	}
	return result
}

func (c *slotsCmd) outputResult() {
	if len(c.redisSlots) == 0 {
		fmt.Fprintln(c.streams.ErrOut, "!! Unable to get any CLUSTER SLOTS data to show..")
		return
	}

	w := tabwriter.NewWriter(c.streams.Out, 6, 4, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w)
	fmt.Fprintln(w, "START\tEND\tROLE\tIP\tPODNAME\tHOST\tREMARKS")

	// Get last podName
	podName := ""
	for k := range c.redisSlots {
		podName = k
	}

	for _, slots := range c.redisSlots[podName] {
		remarksSlots := analyzeSlotsInfo(slots, c.k8sInfo)

		for i, node := range slots.Nodes {
			podInfo := c.k8sInfo.GetPodInfo(node.Addr)
			remarks := podInfo.Info + remarksSlots
			if i == 0 {
				fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%s\t%s\t%s\n",
					slots.Start, slots.End, "master", node.Addr, podInfo.Name, podInfo.Host, remarks)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					".", ".", "repl.", node.Addr, podInfo.Name, podInfo.Host, remarks)
			}
		}
	}
}

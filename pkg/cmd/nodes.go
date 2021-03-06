package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/bjosv/kubectl-rediscluster/pkg/k8s"
	"github.com/bjosv/kubectl-rediscluster/pkg/portforwarder"
	"github.com/bjosv/kubectl-rediscluster/pkg/redisutils"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type nodesCmd struct {
	configFlags *genericclioptions.ConfigFlags
	streams     *genericclioptions.IOStreams
	args        []string
	verbose     bool

	k8sInfo    *k8s.ClusterInfo
	redisInfo  map[string]redisutils.RedisInfo
	redisSlots map[string]redisutils.ClusterSlots
	redisNodes map[string]redisutils.ClusterNodes
	remarks    map[string][]string
	errors     map[string][]string
}

// NewNodesCmd initialize and creates a Cobra command
func NewNodesCmd(streams genericclioptions.IOStreams) *cobra.Command {
	c := &nodesCmd{
		configFlags: genericclioptions.NewConfigFlags(true),
		streams:     &streams,
		k8sInfo:     k8s.NewClusterInfo(),
		redisInfo:   make(map[string]redisutils.RedisInfo),
		redisNodes:  make(map[string]redisutils.ClusterNodes),
		redisSlots:  make(map[string]redisutils.ClusterSlots),
		remarks:     make(map[string][]string),
		errors:      make(map[string][]string),
	}

	cmd := &cobra.Command{
		Use:   "nodes [service-name] [flags]",
		Short: "Show node information of a Redis Cluster",
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
func (c *nodesCmd) Complete(cmd *cobra.Command, args []string) error {
	c.args = args

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (c *nodesCmd) Validate() error {
	if len(c.args) > 1 {
		return fmt.Errorf("maximum 1 service name can be given, got %d", len(c.args))
	}

	return nil
}

// Run the command
func (c *nodesCmd) Run() error {
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
			return fmt.Errorf("%s\n\nPlease provide a service name", err)
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
			redisInfo, clusterNodes, clusterSlots, err := redisutils.QueryRedis(pfwd, namespace, podName, podPort)
			ch <- QueryRedisResult{
				PodName: podName,
				Info:    redisInfo,
				Nodes:   clusterNodes,
				Slots:   clusterSlots,
				Error:   err,
			}
		}(pod.Name, redisutils.RedisPort, ch)
	}

	// Collect results from all pods/redis instances
	for range c.k8sInfo.Pods {
		queryResult := <-ch

		if queryResult.Error != nil {
			pod := queryResult.PodName
			c.remarks[pod] = append(c.remarks[pod], "RedisUnavailable")
			c.errors[pod] = append(c.errors[pod],
				fmt.Sprintf("Failed to get Redis information: %s", queryResult.Error))
		}
		if queryResult.Info != nil {
			c.redisInfo[queryResult.PodName] = queryResult.Info
		}
		if queryResult.Nodes != nil {
			c.redisNodes[queryResult.PodName] = queryResult.Nodes
		}
		if queryResult.Slots != nil {
			c.redisSlots[queryResult.PodName] = queryResult.Slots
		}
	}

	//	Display result
	c.outputResult()

	return nil
}

func (c *nodesCmd) outputResult() {
	// Convert PodInfo to a ordered list
	podList := []k8s.PodInfo{}
	for _, pod := range c.k8sInfo.Pods {
		podList = append(podList, pod)
	}

	// Sort by host and ip
	sort.Slice(podList, func(i, j int) bool {
		if podList[i].Host != podList[j].Host {
			return podList[i].Host < podList[j].Host
		}
		return podList[i].IP < podList[j].IP
	})

	if len(podList) == 0 {
		fmt.Fprintln(c.streams.ErrOut, "!! Unable to get any pod information to show..")
		return
	}

	w := tabwriter.NewWriter(c.streams.Out, 5, 3, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "\t\t\t\t\t\tSLOT\tCLUSTER\t")
	fmt.Fprintln(w, "HOST\tPODNAME\tIP\tROLE\tKEYS\tSLOTS\tRANGES\tSTATE\tUPTIME\tREMARKS")

	for _, p := range podList {
		podName := p.Name
		podIP := p.IP

		nodes := c.redisNodes[podName]
		role := nodes.GetFlagsSelf()

		keys := c.redisInfo[podName]["keys"]

		state := c.redisInfo[podName]["cluster_state"]
		//addr := fmt.Sprintf("%s:%d", p.IP, redisutils.RedisPort)

		uptime := ""
		uptimeSec := c.redisInfo[podName]["uptime_in_seconds"]
		if s, err := strconv.Atoi(uptimeSec); err == nil {
			d := time.Duration(s) * time.Second
			uptime = d.String()
		}

		slots := ""
		slotranges := ""
		if c.redisSlots[podName] != nil {
			s, r := slotsCount(podIP, c.redisSlots[podName])
			slots = strconv.Itoa(s)
			slotranges = strconv.Itoa(r)
		}

		remarks := ""
		for i, info := range c.remarks[podName] {
			if i > 0 {
				remarks += ", "
			}
			remarks += info
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			p.Host, p.Name, p.IP, role, keys, slots, slotranges, state, uptime, remarks)
	}

	// Print errors
	addNewline := true
	for _, p := range podList {
		for _, text := range c.errors[p.Name] {
			if addNewline {
				fmt.Fprintf(w, "\n")
				addNewline = false
			}
			fmt.Fprintf(w, "%s:\t%s\n", p.Name, text)
		}
	}
}

func slotsCount(ip string, slots redisutils.ClusterSlots) (int, int) {
	ep := fmt.Sprintf("%s:%d", ip, redisutils.RedisPort)
	slotsCount := 0
	slotrangesCount := 0
	for _, slot := range slots {
		for _, node := range slot.Nodes {
			if node.Addr == ep {
				slotrangesCount++
				slotsCount += (slot.End - slot.Start + 1)
			}
		}
	}
	return slotsCount, slotrangesCount
}

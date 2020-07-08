package cmd

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/bjosv/kubectl-rediscluster/pkg/k8s"
	"github.com/bjosv/kubectl-rediscluster/pkg/portforwarder"
	"github.com/bjosv/kubectl-rediscluster/pkg/redisutils"
	"github.com/go-redis/redis/v8"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type slotsCmd struct {
	configFlags *genericclioptions.ConfigFlags
	streams     *genericclioptions.IOStreams
	args        []string

	k8sInfo    *k8s.ClusterInfo
	redisSlots []redis.ClusterSlot
}

// Type used when transfering result from portforwarder
type ClusterSlots struct {
	PodName string
	Slots   []redis.ClusterSlot
}

func newSlotsCmd(streams genericclioptions.IOStreams) *cobra.Command {
	c := &slotsCmd{
		configFlags: genericclioptions.NewConfigFlags(true),
		streams:     &streams,
		k8sInfo:     k8s.NewClusterInfo(),
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

	return cmd
}

// Complete sets all information required for the command
func (c *slotsCmd) Complete(cmd *cobra.Command, args []string) error {
	c.args = args

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (c *slotsCmd) Validate() error {
	if len(c.args) != 1 {
		return fmt.Errorf("a single service name is required, got %d", len(c.args))
	}

	return nil
}

func (c *slotsCmd) Run() error {

	serviceName := c.args[0]

	namespace, err := currentNamespace(c.configFlags)
	if err != nil {
		return err
	}

	restConfig, err := c.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	err = getK8sInfo(restConfig, serviceName, namespace, c.k8sInfo)
	if err != nil {
		return err
	}

	//Verbose
	//pfwd := portforwarder.New(restConfig, c.streams.Out, c.streams.ErrOut)
	pfwd := portforwarder.New(restConfig, nil, nil)

	// Silence K8s errors, like connection refuse
	logKubeError := func(err error) {}
	runtime.ErrorHandlers = []func(error){logKubeError}

	// Get CLUSTER SLOTS from all Redis pods
	ch := make(chan ClusterSlots)
	for _, pod := range c.k8sInfo.Pods {
		go func(podName string, podPort int, ch chan ClusterSlots) {
			clusterSlots, err := redisutils.GetClusterSlots(pfwd, namespace, podName, podPort)
			if err != nil {
				fmt.Printf("Failed to get Redis Cluster slot information for pod=%s: %v\n",
					podName, err)
			}
			ch <- ClusterSlots{
				PodName: podName,
				Slots:   clusterSlots,
			}
		}(pod.Name, 6379, ch)
	}

	// Collect result from CLUSTER SLOTS
	for range c.k8sInfo.Pods {
		select {
		case clusterSlots := <-ch:
			if clusterSlots.Slots != nil {
				if c.redisSlots == nil {
					c.redisSlots = clusterSlots.Slots
				} else {
					// Merge slots

				}
			} else {
				fmt.Printf("Slots data missing from %s\n", clusterSlots.PodName)
			}
		}
	}

	//	Display result
	err = c.outputResult()
	if err != nil {
		return err
	}

	return nil
}

func getK8sInfo(restConfig *rest.Config, serviceName string, namespace string, k8sInfo *k8s.ClusterInfo) error {

	clientset := kubernetes.NewForConfigOrDie(restConfig)

	// Check that the service exists, needed to get the pod label selector
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get service/%s in namespace/%s: %v\n", serviceName, namespace, err)
	}

	// Get pod information from the Endpoint resource
	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get endpoints/%s in namespace/%s: %v\n", serviceName, namespace, err)
	}
	k8sInfo.AddPodEndpoints(endpoints)

	// Get pods that matches the Service label selector
	labelMap := service.Spec.Selector
	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), options)
	if err != nil {
		return fmt.Errorf("Failed to list pods in namespace/%s: %v\n", namespace, err)
	}
	k8sInfo.UpdatePods(pods)

	return nil
}

func currentNamespace(configFlags *genericclioptions.ConfigFlags) (string, error) {
	kubeConfig := configFlags.ToRawKubeConfigLoader()
	namespace, _, err := kubeConfig.Namespace()
	return namespace, err
}

func (c *slotsCmd) outputResult() error {
	if len(c.redisSlots) == 0 {
		fmt.Fprintln(c.streams.ErrOut, "!! Unable to get any CLUSTER SLOTS data to show..")
		return nil
	}

	w := tabwriter.NewWriter(c.streams.Out, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintln(w, "START\tEND\tMASTER\tREPLICA\tPODNAME\tNODE\tINFO")
	for _, slot := range c.redisSlots {
		for i, node := range slot.Nodes {
			addr := node.Addr
			podInfo := c.k8sInfo.GetPodInfo(addr)
			if i == 0 {
				fmt.Fprintln(w, fmt.Sprintf("%d\t%d\t%s\t%s\t%s\t%s\t%s",
					slot.Start, slot.End, addr, "", podInfo.Name, podInfo.Node, podInfo.Info))
			} else {
				fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s",
					"", "", "", addr, podInfo.Name, podInfo.Node, podInfo.Info))
			}
		}
	}
	w.Flush()
	return nil
}

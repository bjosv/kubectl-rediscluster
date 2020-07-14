package cmd

import (
	"context"
	"fmt"

	"github.com/bjosv/kubectl-rediscluster/pkg/k8s"
	"github.com/bjosv/kubectl-rediscluster/pkg/redisutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Type used when transferring result from portforwarder
type QueryRedisResult struct {
	PodName string
	Info    redisutils.ClusterInfo
	Nodes   redisutils.ClusterNodes
	Slots   redisutils.ClusterSlots
	Error   error
}

func getK8sInfo(restConfig *rest.Config, serviceName string, namespace string, k8sInfo *k8s.ClusterInfo) error {
	clientset := kubernetes.NewForConfigOrDie(restConfig)

	// Check that the service exists, needed to get the pod label selector
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get service/%s in namespace/%s: %v",
			serviceName, namespace, err)
	}

	// Get pod information from the Endpoint resource
	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get endpoints/%s in namespace/%s: %v",
			serviceName, namespace, err)
	}
	k8sInfo.AddPodEndpoints(endpoints)

	// Get pods that matches the Service label selector
	labelMap := service.Spec.Selector
	if len(labelMap) == 0 {
		return fmt.Errorf("the service %s/%s has an empty pod selector",
			namespace, serviceName)
	}

	var timeout int64 = 2
	options := metav1.ListOptions{
		LabelSelector:  labels.SelectorFromSet(labelMap).String(),
		TimeoutSeconds: &timeout,
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), options)
	if err != nil {
		return fmt.Errorf("failed to list pods in namespace/%s: %v", namespace, err)
	}
	k8sInfo.UpdatePods(pods)

	return nil
}

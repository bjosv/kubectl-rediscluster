package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CurrentNamespace get the namespace in use
func CurrentNamespace(configFlags *genericclioptions.ConfigFlags) (string, error) {
	kubeConfig := configFlags.ToRawKubeConfigLoader()
	namespace, _, err := kubeConfig.Namespace()
	return namespace, err
}

// FindServiceUsingPort tries to find a service using a specific port
func FindServiceUsingPort(restConfig *rest.Config, namespace string, port int) (string, error) {
	clientset := kubernetes.NewForConfigOrDie(restConfig)

	options := metav1.ListOptions{}
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), options)
	if err != nil {
		return "", fmt.Errorf("Failed to list services in namespace/%s: %v\n", namespace, err)
	}

	for _, item := range services.Items {
		for _, p := range item.Spec.Ports {
			if int(p.Port) == port {
				return item.ObjectMeta.Name, nil
			}
		}
	}
	return "", fmt.Errorf("Could not find a service using port=%d in namespace/%s\n", port, namespace)
}

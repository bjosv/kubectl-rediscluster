package k8s

import (
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
)

type PodInfo struct {
	Name      string
	IP        string
	Node      string
	Restarts  int
	StartTime string
	Info      string
}

type ClusterInfo struct {
	Pods map[string]PodInfo
}

func NewClusterInfo() *ClusterInfo {
	return &ClusterInfo{
		Pods: make(map[string]PodInfo),
	}
}

func (c *ClusterInfo) GetPodInfo(podAddress string) PodInfo {
	parts := strings.Split(podAddress, ":")
	if len(parts) > 1 {
		return c.Pods[parts[0]]
	}
	return c.Pods[podAddress]
}

// TODO: handle merger of pod info
func (c *ClusterInfo) AddPodEndpoints(endpoints *v1.Endpoints) {
	for _, eps := range endpoints.Subsets {
		for _, epPort := range eps.Ports {
			_ = epPort.Port

			for _, epAddress := range eps.Addresses {
				if epAddress.TargetRef.Kind == "Pod" {
					ip := epAddress.IP
					p := PodInfo{
						Name: epAddress.TargetRef.Name,
						IP:   ip,
						Node: *epAddress.NodeName,
					}
					c.Pods[ip] = p
					//fmt.Printf("> Pod added: %s\n", p.Name)
				}
			}
		}
	}
}

func (c *ClusterInfo) UpdatePods(podList *v1.PodList) {

	//fmt.Printf("PodList => %+v\n", podList)
	for _, pod := range podList.Items {
		ip := pod.Status.PodIP
		//fmt.Printf("  POD => %+v\n", pod)

		if p, ok := c.Pods[ip]; ok {
			p.StartTime = pod.Status.StartTime.String()
			for _, container := range pod.Status.ContainerStatuses {
				//fmt.Printf("  Container => %+v\n", container.Name)
				p.Restarts = int(container.RestartCount)
			}
			c.Pods[ip] = p
			//fmt.Printf("> Pod updated: %s\n", p.Name)
		} else {
			fmt.Fprintf(os.Stderr, "Selector matches %s (%s), but its not included in the Endpoint resource\n", pod.ObjectMeta.Name, ip)
			p := PodInfo{
				Name: pod.ObjectMeta.Name,
				IP:   ip,
				Node: pod.Spec.NodeName,
				Info: "Endpoint data missing",
			}
			c.Pods[ip] = p
			//fmt.Printf("> Pod added (missing): %s\n", p.Name)
		}
	}
	//fmt.Println("Update done")
}

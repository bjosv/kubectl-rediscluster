package portforwarder

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwarder struct {
	// RestConfig is the kubernetes config
	restConfig *rest.Config
	// Out think, os.Stdout
	out io.Writer
	// ErrOut think, os.Stderr
	errOut io.Writer
}

func New(restConfig *rest.Config, out io.Writer, errOut io.Writer) *PortForwarder {
	return &PortForwarder{restConfig, out, errOut}
}

func (p *PortForwarder) ForwardPort(podNamespace string, podName string, localPort int, podPort int, stopCh <-chan struct{}, readyCh chan struct{}) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", podNamespace, podName)
	hostIP := strings.TrimLeft(p.restConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(p.restConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})

	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, podPort)}, stopCh, readyCh, p.out, p.errOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func GetAvailableLocalPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

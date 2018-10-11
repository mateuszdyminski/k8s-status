package runner

import (
	"context"
	"fmt"

	"github.com/mateuszdyminski/k8s-status/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

// healthzChecker is secure healthz checker
type healthzChecker struct {
	*KubeChecker
}

// KubeAPIServerHealth creates a checker for the kubernetes API server
func KubeAPIServerHealth(config KubeConfig) Checker {
	checker := &healthzChecker{}
	kubeChecker := &KubeChecker{
		name:    "kube-apiserver",
		checker: checker.testHealthz,
		client:  config.Client,
	}
	checker.KubeChecker = kubeChecker
	return kubeChecker
}

// testHealthz executes a test by using k8s API
func (h *healthzChecker) testHealthz(ctx context.Context, client *kube.Clientset) error {
	_, err := client.CoreV1().ComponentStatuses().Get("scheduler", metav1.GetOptions{})
	return err
}

// KubeletHealth creates a checker for the kubernetes kubelet component
func KubeletHealth(addr string) Checker {
	return NewHTTPHealthzChecker("kubelet", fmt.Sprintf("%v/healthz", addr), kubeHealthz)
}

// NodesStatusHealth creates a checker that reports a number of ready kubernetes nodes
func NodesStatusHealth(config KubeConfig, nodesReadyThreshold int) Checker {
	return NewNodesStatusChecker(config, nodesReadyThreshold)
}

// EtcdHealth creates a checker that checks health of etcd
func EtcdHealth(cfg *config.ETCDConfig) (Checker, error) {
	const name = "etcd-healthz"

	transport, err := cfg.NewHTTPTransport()
	if err != nil {
		return nil, err
	}
	createChecker := func(addr string) (Checker, error) {
		endpoint := fmt.Sprintf("%v/health", addr)
		return NewHTTPHealthzCheckerWithTransport(name, endpoint, transport, config.EtcdChecker), nil
	}
	var checkers []Checker
	for _, endpoint := range cfg.Endpoints {
		checker, err := createChecker(endpoint)
		if err != nil {
			return nil, err
		}
		checkers = append(checkers, checker)
	}
	return &compositeChecker{name, checkers}, nil
}

func (_ noopChecker) Name() string                    { return "noop" }
func (_ noopChecker) Check(context.Context, Reporter) {}

// noopChecker is a checker that does nothing
type noopChecker struct{}

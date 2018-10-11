package runner

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
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

// KubeEtcdHealth creates a checker that checks health of etcd
func KubeEtcdHealth(config KubeConfig) Checker {
	checker := &healthzChecker{}
	kubeChecker := &KubeChecker{
		name:    "etcd",
		checker: checker.testHealthz,
		client:  config.Client,
	}
	checker.KubeChecker = kubeChecker
	return kubeChecker
}

// testHealthz executes a test by using k8s API
func (h *healthzChecker) testEtcdHealthz(ctx context.Context, client *kube.Clientset) error {
	res, err := client.CoreV1().ComponentStatuses().List(metav1.ListOptions{Limit: 100})

	for _, item := range res.Items {
		log.Info().Msgf("%+v", item)
	}

	return err
}

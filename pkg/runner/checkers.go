package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/mateuszdyminski/k8s-status/pkg/config"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

// healthzChecker is secure healthz checker
type healthzChecker struct {
	*KubeChecker
}

// KubeClusterConfig creates a checker that checks health of etcd
func KubeClusterConfig(config KubeConfig, cfg *config.Config) Checker {
	return clusterConfig(config, cfg)
}

// KubeEtcdHealth creates a checker that checks health of etcd
func KubeEtcdHealth(config KubeConfig) Checker {
	return componentServerHealth(config, "etcd")
}

// KubeSchedulerHealth creates a checker that checks health of scheduler
func KubeSchedulerHealth(config KubeConfig) Checker {
	return componentServerHealth(config, "scheduler")
}

// KubeControllerManagerHealth creates a checker that checks health of controller manager
func KubeControllerManagerHealth(config KubeConfig) Checker {
	return componentServerHealth(config, "controller-manager")
}

// NodesStatusHealth creates a checker that reports a number of ready kubernetes nodes
func NodesStatusHealth(config KubeConfig, nodesReadyThreshold int) Checker {
	return NewNodesStatusChecker(config, nodesReadyThreshold)
}

// KubeAPIServerHealth creates a checker for the kubernetes API server
func componentServerHealth(config KubeConfig, componentName string) Checker {
	checker := &healthzChecker{}
	kubeChecker := &KubeChecker{
		name:    componentName,
		checker: checker.testHealthz(componentName),
		client:  config.Client,
	}
	checker.KubeChecker = kubeChecker
	return kubeChecker
}

// testHealthz executes a test by using k8s API
func (h *healthzChecker) testHealthz(componentName string) KubeStatusChecker {
	return h.testComponentHeathz(componentName)
}

func clusterConfig(kubeConfig KubeConfig, cfg *config.Config) Checker {
	checker := &healthzChecker{}
	kubeChecker := &KubeChecker{
		name:    cfg.ConfigCheckerConfigName,
		checker: checker.clusterConfig(cfg),
		client:  kubeConfig.Client,
	}
	checker.KubeChecker = kubeChecker
	return kubeChecker
}

func (h *healthzChecker) clusterConfig(cfg *config.Config) KubeStatusChecker {
	return func(ctx context.Context, client *kube.Clientset) (interface{}, error) {
		res, err := client.CoreV1().ConfigMaps(cfg.ConfigCheckerNamespace).Get(cfg.ConfigCheckerConfigName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		return res.Data, nil
	}
}

// testHealthz executes a test by using k8s API
func (h *healthzChecker) testComponentHeathz(componentName string) KubeStatusChecker {
	return func(ctx context.Context, client *kube.Clientset) (interface{}, error) {
		res, err := client.CoreV1().ComponentStatuses().List(metav1.ListOptions{LabelSelector: fmt.Sprintf("component=%s", componentName), Limit: 100})
		if err != nil {
			return nil, err
		}

		healthy := true
		conditions := make([]interface{}, 0, 0)
		for _, item := range res.Items {
			if strings.Contains(item.GetName(), componentName) {
				for _, condition := range item.Conditions {
					conditions = append(conditions, condition)
					if condition.Type != v1.ComponentHealthy {
						healthy = false
						continue
					}
					if condition.Status != v1.ConditionTrue {
						healthy = false
						continue
					}
				}
			}
		}

		if len(conditions) == 0 {
			return nil, fmt.Errorf("no component: %s found on Kubernetes cluster", componentName)
		}

		if !healthy {
			return nil, fmt.Errorf("component: %s is not healthy", componentName)
		}

		return conditions, nil
	}
}

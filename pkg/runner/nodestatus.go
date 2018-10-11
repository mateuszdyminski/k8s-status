package runner

import (
	"context"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// NewNodesStatusChecker returns a Checker that tests kubernetes nodes availability
func NewNodesStatusChecker(config KubeConfig, nodesReadyThreshold int) Checker {
	return &nodesStatusChecker{
		client:              config.Client.CoreV1(),
		nodesReadyThreshold: nodesReadyThreshold,
	}
}

// nodesStatusChecker tests and reports health failures in kubernetes
// nodes availability
type nodesStatusChecker struct {
	client              corev1.CoreV1Interface
	nodesReadyThreshold int
}

// Name returns the name of this checker
func (r *nodesStatusChecker) Name() string { return NodesStatusCheckerID }

// Check validates the status of kubernetes components
func (r *nodesStatusChecker) Check(ctx context.Context, reporter Reporter) {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
		FieldSelector: fields.Everything().String(),
	}
	statuses, err := r.client.Nodes().List(listOptions)
	if err != nil {
		reason := "failed to query nodes"
		reporter.Add(NewProbeFromErr(r.Name(), reason, err))
		return
	}
	var nodesReady int
	for _, item := range statuses.Items {
		for _, condition := range item.Status.Conditions {
			if condition.Type != v1.NodeReady {
				continue
			}
			if condition.Status == v1.ConditionTrue {
				nodesReady++
				continue
			}
		}
	}

	if nodesReady < r.nodesReadyThreshold {
		reporter.Add(&Probe{
			Checker: r.Name(),
			Status:  ProbeFailed,
			Error: fmt.Sprintf("Not enough ready nodes: %v (threshold %v)",
				nodesReady, r.nodesReadyThreshold),
		})
	} else {
		reporter.Add(&Probe{
			Checker: r.Name(),
			Status:  ProbeRunning,
		})
	}
}

// NewNodeStatusChecker returns a Checker that validates availability
// of a single kubernetes node
func NewNodeStatusChecker(config KubeConfig, nodeName string) Checker {
	nodeLister := kubeNodeLister{client: config.Client.CoreV1()}
	return &nodeStatusChecker{
		nodeLister: nodeLister,
		nodeName:   nodeName,
	}
}

// NewNodeStatusChecker returns a Checker that validates availability
// of a single kubernetes node
type nodeStatusChecker struct {
	nodeLister
	nodeName string
}

// Name returns the name of this checker
func (r *nodeStatusChecker) Name() string { return NodeStatusCheckerID }

// Check validates the status of kubernetes components
func (r *nodeStatusChecker) Check(ctx context.Context, reporter Reporter) {
	options := metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
		FieldSelector: fields.SelectorFromSet(fields.Set{"metadata.name": r.nodeName}).String(),
	}
	nodes, err := r.nodeLister.Nodes(options)
	if err != nil {
		reporter.Add(NewProbeFromErr(r.Name(), err.Error(), err))
		return
	}

	if len(nodes.Items) != 1 {
		reporter.Add(NewProbeFromErr(r.Name(), "",
			fmt.Errorf("node %q not found", r.nodeName)))
		return
	}

	node := nodes.Items[0]
	var failureCondition *v1.NodeCondition
	for _, condition := range node.Status.Conditions {
		if condition.Type != v1.NodeReady {
			continue
		}
		if condition.Status != v1.ConditionTrue && node.Name == r.nodeName {
			failureCondition = &condition
			break
		}
	}

	if failureCondition == nil {
		reporter.Add(&Probe{
			Checker: r.Name(),
			Status:  ProbeRunning,
		})
		return
	}

	reporter.Add(&Probe{
		Checker:  r.Name(),
		Status:   ProbeFailed,
		Severity: ProbeWarning,
		Detail:   formatCondition(*failureCondition),
		Error:    "Node is not ready",
	})
}

type nodeLister interface {
	Nodes(metav1.ListOptions) (*v1.NodeList, error)
}

func (r kubeNodeLister) Nodes(options metav1.ListOptions) (*v1.NodeList, error) {
	nodes, err := r.client.Nodes().List(options)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes. err: %s", err)
	}
	return nodes, nil
}

type kubeNodeLister struct {
	client corev1.CoreV1Interface
}

func formatCondition(condition v1.NodeCondition) string {
	if condition.Message != "" {
		return fmt.Sprintf("%v (%v)", condition.Reason, condition.Message)
	}
	return condition.Reason
}

const (
	// NodeStatusCheckerID identifies the checker that detects whether a node is not ready
	NodeStatusCheckerID = "nodestatus"
	// NodesStatusCheckerID identifies the checker that validates node availability in a cluster
	NodesStatusCheckerID = "nodesstatus"
)

package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	kube "k8s.io/client-go/kubernetes"
)

// KubeConfig defines Kubernetes access configuration
type KubeConfig struct {
	// Client is the initialized Kubernetes client
	Client *kube.Clientset
}

// kubeHealthz is httpResponseChecker that interprets health status of common kubernetes services.
func kubeHealthz(response io.Reader) error {
	payload, err := ioutil.ReadAll(response)
	if err != nil {
		return err
	}
	if !bytes.Equal(payload, []byte("ok")) {
		return fmt.Errorf("unexpected healthz response: %s", payload)
	}
	return nil
}

// KubeStatusChecker is a function that can check status of kubernetes services.
type KubeStatusChecker func(ctx context.Context, client *kube.Clientset) error

// KubeChecker implements Checker that can check and report problems
// with kubernetes services.
type KubeChecker struct {
	name    string
	checker KubeStatusChecker
	client  *kube.Clientset
}

// Name returns the name of this checker
func (r *KubeChecker) Name() string { return r.name }

// Check runs the wrapped kubernetes service checker function and reports
// status to the specified reporter
func (r *KubeChecker) Check(ctx context.Context, reporter Reporter) {
	err := r.checker(ctx, r.client)
	if err != nil {
		reporter.Add(NewProbeFromErr(r.name, noErrorDetail, err))
		return
	}
	reporter.Add(&Probe{
		Checker: r.name,
		Status:  ProbeRunning,
	})
}

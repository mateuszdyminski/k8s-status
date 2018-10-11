package runner

import (
	"context"
	"fmt"

	"github.com/mateuszdyminski/k8s-status/pkg/config"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Runner stores configured checkers and provides interface to configure
// and run them
type Runner struct {
	Checkers
}

// NewRunner creates Runner with checks configured using provided options
func NewRunnerWithCfg(cfg *config.Config) (*Runner, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kubeConfig := KubeConfig{Client: clientset}

	runner := &Runner{}
	runner.AddChecker(KubeEtcdHealth(kubeConfig))
	runner.AddChecker(KubeAPIServerHealth(kubeConfig))
	runner.AddChecker(NodesStatusHealth(kubeConfig, cfg.KubeNodesReadyThreshold))
	return runner, nil
}

// Run runs all checks successively and reports general cluster status
func (c *Runner) Run(ctx context.Context) *FinalProbe {
	var probes Probes

	for _, c := range c.Checkers {
		log.Info().Msgf("running checker %s", c.Name())
		c.Check(ctx, &probes)
	}

	return finalHealth(probes)
}

// finalHealth aggregates statuses from all probes into one summarized health status
func finalHealth(probes Probes) *FinalProbe {
	var errors []string
	status := ProbeRunning

	for _, probe := range probes {
		switch probe.Status {
		case ProbeRunning:
			errors = append(errors, fmt.Sprintf("Check %s: OK", probe.Checker))
		default:
			status = ProbeFailed
			errors = append(errors, fmt.Sprintf("Check %s: %s", probe.Checker, probe.Error))
		}
	}

	clusterHealth := FinalProbe{
		Status: status,
		Errors: errors,
	}

	log.Info().Msgf("cluster new health: %#v", clusterHealth)

	return &clusterHealth
}

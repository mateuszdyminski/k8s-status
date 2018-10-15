package config

import "github.com/kelseyhightower/envconfig"

// Config holds configuration.
type Config struct {
	HTTPPort                   int
	GracefulShutdownTimeout    int
	GracefulShutdownExtraSleep int
	Debug                      bool

	// Kubernetes nodes config
	KubeNodesReadyThreshold int

	// Config checker
	ConfigCheckerNamespace  string
	ConfigCheckerConfigName string
}

// LoadConfig loads config from env vars.
func LoadConfig() (*Config, error) {
	var c Config
	err := envconfig.Process("k8status", &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

package main

import (
	"github.com/mateuszdyminski/k8s-status/pkg/config"
	"github.com/mateuszdyminski/k8s-status/pkg/runner"
	"github.com/mateuszdyminski/k8s-status/pkg/server"
	"github.com/mateuszdyminski/k8s-status/pkg/signals"
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Msgf("can't load config file. err: %s", err)
	}

	runner, err := runner.NewRunnerWithCfg(cfg)
	if err != nil {
		log.Fatal().Msgf("can't create health runner. err: %s", err)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ctx := signals.SetupSignalContext()
	server.ListenAndServe(ctx, runner, cfg)
}

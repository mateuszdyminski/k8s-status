package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/mateuszdyminski/k8s-status/pkg/config"
	"github.com/mateuszdyminski/k8s-status/pkg/runner"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	healthy int32 = 1
	ready   int32 = 1
)

type Server struct {
	mux    *mux.Router
	runner *runner.Runner
}

func NewServer(cfg *config.Config, runner *runner.Runner, options ...func(*Server)) *Server {
	s := &Server{runner: runner, mux: mux.NewRouter()}

	for _, f := range options {
		f(s)
	}

	// metrics
	s.mux.Handle("/metrics", promhttp.Handler())

	// register general handlers
	s.mux.HandleFunc("/healthz", s.healthz)
	s.mux.HandleFunc("/readyz", s.readyz)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", runtime.Version())

	s.mux.ServeHTTP(w, r)
}

func ListenAndServe(cancelCtx context.Context, runner *runner.Runner, cfg *config.Config) {
	inst := NewInstrument()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      inst.Wrap(NewServer(cfg, runner)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  15 * time.Second,
	}

	// run server in background
	go func() {
		log.Info().Msgf("HTTP Server started at port: %d", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server crashed")
		}
	}()

	// wait for SIGTERM or SIGINT
	<-cancelCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.GracefulShutdownTimeout)*time.Second)
	defer cancel()

	// all calls to /healthz and /readyz will fail from now on
	atomic.StoreInt32(&healthy, 0)
	atomic.StoreInt32(&ready, 0)

	time.Sleep(time.Duration(int64(cfg.GracefulShutdownExtraSleep) * int64(time.Second)))

	log.Info().Msgf("Shutting down HTTP server with timeout: %v", time.Duration(cfg.GracefulShutdownTimeout)*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP server graceful shutdown failed")
	} else {
		log.Info().Msg("HTTP server stopped")
	}
}

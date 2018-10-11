package server

import (
	"context"
	"encoding/json"
	"net/http"
)

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	clusterHealth := s.runner.Run(context.TODO())

	data, err := json.Marshal(clusterHealth)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

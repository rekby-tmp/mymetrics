package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	endpoint string
	storage  Storage
	mux      http.ServeMux
	server   *http.Server
}

func NewServer(endpoint string, storage Storage) *Server {
	s := &Server{
		endpoint: endpoint,
		storage:  storage,
	}
	s.mux.HandleFunc("/update/counter/", s.handleCounter)
	s.mux.HandleFunc("/update/gauge/", s.handleGauge)
	s.server = &http.Server{
		Addr:    s.endpoint,
		Handler: &s.mux,
	}
	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) handleCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unexpected method: %q", r.Method), http.StatusBadRequest)
		return
	}

	name, valS, err := getMetricNameValue(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get metric value error: %v", err), http.StatusBadRequest)
		return
	}

	val, err := strconv.ParseInt(valS, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Parse int value error: %v", err), http.StatusBadRequest)
		return
	}

	err = s.storage.StoreCounter(name, val)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to store counter value: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unexpected method: %q", r.Method), http.StatusBadRequest)
		return
	}

	name, valS, err := getMetricNameValue(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get metric value error: %v", err), http.StatusBadRequest)
		return
	}

	val, err := strconv.ParseFloat(valS, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Parse float value error: %v", err), http.StatusBadRequest)
		return
	}

	err = s.storage.StoreGauge(name, val)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to store gauge value: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getMetricNameValue(r *http.Request) (name, value string, _ error) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("bad parts in path for extrace metric name and balue: %v", len(parts))
	}

	return parts[len(parts)-2], parts[len(parts)-1], nil
}

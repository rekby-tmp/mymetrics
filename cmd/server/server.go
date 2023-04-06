package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	endpoint string
	storage  Storage
	mux      http.ServeMux
}

func NewServer(endpoint string, storage Storage) *Server {
	s := &Server{
		endpoint: endpoint,
		storage:  storage,
	}
	s.mux.HandleFunc("/update/counter/", s.handleGauge)
	s.mux.HandleFunc("/update/gauge/", s.handleGauge)
	return s
}

func (s *Server) Start() error {
	httpServer := http.Server{
		Addr:    s.endpoint,
		Handler: &s.mux,
	}
	return httpServer.ListenAndServe()
}

func (s *Server) handleCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unexpected method: %q", r.Method), http.StatusBadRequest)
		return
	}

	name, valS, err := getMetricNameValue(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get metric value error: %w", err), http.StatusBadRequest)
		return
	}

	val, err := strconv.ParseInt(valS, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Parse int value error: %w", err), http.StatusBadRequest)
		return
	}

	s.storage.StoreCounter(name, val)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unexpected method: %q", r.Method), http.StatusBadRequest)
		return
	}

	name, valS, err := getMetricNameValue(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Get metric value error: %w", err), http.StatusBadRequest)
		return
	}

	val, err := strconv.ParseFloat(valS, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Parse float value error: %w", err), http.StatusBadRequest)
		return
	}

	s.storage.StoreGauge(name, val)
	w.WriteHeader(http.StatusOK)
}

func getMetricNameValue(r *http.Request) (name, value string, _ error) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("bad parts in path for extrace metric name and balue: %v", len(parts))
	}

	return parts[len(parts)-2], parts[len(parts)-1], nil
}

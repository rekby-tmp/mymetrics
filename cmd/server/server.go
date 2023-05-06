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
	s.mux.HandleFunc("/update/", s.updateHandler)
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

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, fmt.Sprintf("Unexpected method: %q", r.Method), http.StatusBadRequest)
		return
	}

	valType, name, valS, err := getMetricTypeNameValue(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Path not found: %v", err), http.StatusNotFound)
		return
	}

	switch valType {
	case "gauge":
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
	case "counter":
		val, err := strconv.ParseInt(valS, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Parse float value error: %v", err), http.StatusBadRequest)
			return
		}
		err = s.storage.StoreCounter(name, val)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to store counter value: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("unknown value type: %q", valType), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getMetricTypeNameValue(r *http.Request) (valType, name, value string, _ error) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return "", "", "", fmt.Errorf("bad parts in path for extrace metric name and balue: %v", len(parts))
	}

	return parts[1], parts[2], parts[3], nil
}

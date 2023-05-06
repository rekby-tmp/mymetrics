package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Endpoint string
	storage  Storage
	r        *chi.Mux
	server   *http.Server
}

func NewServer(endpoint string, storage Storage) *Server {
	s := &Server{
		Endpoint: endpoint,
		storage:  storage,
		r:        chi.NewRouter(),
	}
	s.r.Get("/", s.listMetrics)
	s.r.Post("/update/{valType}/{name}/{value}", s.updateMetric)
	s.r.Get("/value/{valType}/{name}", s.getMetric)
	s.server = &http.Server{
		Addr:    s.Endpoint,
		Handler: s.r,
	}
	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) getMetric(w http.ResponseWriter, r *http.Request) {
	valType := chi.URLParam(r, "valType")
	name := chi.URLParam(r, "name")

	res, err := s.storage.Get(name, MetricType(valType))
	switch {
	case errors.Is(err, errUnknownMetricType):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, errNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		// pass
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, res)
}

func (s *Server) listMetrics(w http.ResponseWriter, r *http.Request) {
	list, err := s.storage.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}
	for t, names := range list {
		_, _ = buf.WriteString(string(t))
		_, _ = buf.WriteString(":\n")

		for _, name := range names {
			_, _ = buf.WriteString(name)
			_ = buf.WriteByte('\n')
		}
		_ = buf.WriteByte('\n')
	}

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, buf)
}

func (s *Server) updateMetric(w http.ResponseWriter, r *http.Request) {
	valType := chi.URLParam(r, "valType")
	name := chi.URLParam(r, "name")
	valS := chi.URLParam(r, "value")

	err := s.storage.Store(name, MetricType(valType), valS)
	switch {
	case errors.Is(err, errUnknownMetricType) || errors.Is(err, errBadValue):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		// pass
	}
	w.WriteHeader(http.StatusOK)
}

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rekby-tmp/mymetrics/internal/common"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Server struct {
	Endpoint string
	storage  Storage
	r        *chi.Mux
	server   *http.Server
	logger   *zap.Logger
}

func NewServer(endpoint string, storage Storage, logger *zap.Logger) *Server {
	s := &Server{
		Endpoint: endpoint,
		storage:  storage,
		r:        chi.NewRouter(),
		logger:   logger,
	}
	s.r.Get("/", s.listMetrics)
	s.r.Post("/update/", s.updateJson)
	s.r.Post("/update/{valType}/{name}/{value}", s.updateMetric)
	s.r.Post("/value/", s.getValueJson)
	s.r.Get("/value/{valType}/{name}", s.getMetric)

	var handler http.Handler = s.r
	handler = WithLogging(logger, handler)

	s.server = &http.Server{
		Addr:    s.Endpoint,
		Handler: handler,
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

	res, err := s.storage.Get(name, common.MetricType(valType))
	switch {
	case errors.Is(err, ErrUnknownMetricType):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		// pass
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, fmt.Sprint(res))
}

func (s *Server) listMetrics(w http.ResponseWriter, _ *http.Request) {
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

	val, err := ParseMetricValue(common.MetricType(valType), valS)
	switch {
	case errors.Is(err, ErrUnknownMetricType):
		// pass
	case err != nil:
		err = fmt.Errorf("failed to parse value %q: %w", err, ErrBadValue)
	default:
		err = s.storage.Store(name, common.MetricType(valType), val)
	}

	switch {
	case errors.Is(err, ErrUnknownMetricType) || errors.Is(err, ErrBadValue):
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

func (s *Server) updateJson(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != common.JsonType {
		http.Error(w, "Accept "+common.JsonType+" content type only", http.StatusBadRequest)
		return
	}

	var m common.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		s.logger.Warn("failed to read request content", zap.Error(err))
		http.Error(w, "failed to read request content", http.StatusInternalServerError)
	}

	var storeErr error
	switch m.MType {
	case common.MetricTypeCounter:
		if m.Delta == nil {
			http.Error(w, "data has no Delta value", http.StatusBadRequest)
			return
		}
		var newVal any
		newVal, storeErr = s.storage.StoreAndGet(m.ID, common.MetricTypeCounter, *m.Delta)
		m.Delta = &(&[1]int64{newVal.(int64)})[0]
	case common.MetricTypeGauge:
		if m.Value == nil {
			http.Error(w, "data has no Value value", http.StatusBadRequest)
			return
		}
		var newVal any
		newVal, storeErr = s.storage.StoreAndGet(m.ID, common.MetricTypeGauge, *m.Value)
		m.Value = &(&[1]float64{newVal.(float64)})[0]
	default:
		s.logger.Warn("unknown metric type", zap.Stringer("metric_type", m.MType))
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}
	if storeErr != nil {
		s.logger.Error("failed to store value", zap.Error(storeErr))
		http.Error(w, "failed store value", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", common.JsonType)
	err = json.NewEncoder(w).Encode(m)
	if err != nil {
		s.logger.Error("failed to encore json response", zap.Error(err))
		http.Error(w, "failed store value", http.StatusInternalServerError)
	}
}

func (s *Server) getValueJson(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != common.JsonType {
		http.Error(w, "Accept "+common.JsonType+" content type only", http.StatusBadRequest)
		return
	}

	var m common.Metric
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		s.logger.Warn("failed to read request content", zap.Error(err))
		http.Error(w, "failed to read request content", http.StatusInternalServerError)
	}

	if m.Delta != nil || m.Value != nil {
		s.logger.Warn("bad request - non empty value or delta field", zap.Any("req-object", m))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mType := common.MetricType(m.MType)
	val, err := s.storage.Get(m.ID, mType)
	switch {
	case errors.Is(err, ErrUnknownMetricType):
		s.logger.Warn("unknown metric type", zap.Stringer("metric_type", m.MType))
		http.Error(w, "bad metric type", http.StatusBadRequest)
		return
	case errors.Is(err, ErrNotFound):
		s.logger.Warn("metric not found", zap.Stringer("metric_type", m.MType), zap.String("name", m.ID))
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	case err != nil:
		s.logger.Warn("fail to get metric", zap.Stringer("metric_type", m.MType), zap.String("name", m.ID), zap.Error(err))
		http.Error(w, "fail to get metric", http.StatusInternalServerError)
		return
	default:
		// pass
	}

	res := common.Metric{
		ID:    m.ID,
		MType: m.MType,
	}
	switch mType {
	case common.MetricTypeCounter:
		val := val.(int64)
		res.Delta = &val
	case common.MetricTypeGauge:
		val := val.(float64)
		res.Value = &val
	default:
		s.logger.Error("unexpected metric type", zap.Stringer("metric_type", m.MType), zap.String("name", m.ID))
		http.Error(w, "unexpected metric type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", common.JsonType)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		s.logger.Error("failed to encode result", zap.Stringer("metric_type", m.MType), zap.String("name", m.ID), zap.Error(err))
		http.Error(w, "failed to encode result", http.StatusInternalServerError)
		return
	}
}

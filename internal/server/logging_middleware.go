package server

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ResponseWriterWithStat struct {
	http.ResponseWriter
	Size       int
	StatusCode int
}

func (w *ResponseWriterWithStat) Write(b []byte) (int, error) {
	bytes, err := w.ResponseWriter.Write(b)
	w.Size += bytes
	return bytes, err
}

func (w *ResponseWriterWithStat) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func WithLogging(logger *zap.Logger, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		calcSize := ResponseWriterWithStat{ResponseWriter: w, StatusCode: http.StatusOK}

		h.ServeHTTP(&calcSize, r)

		duration := time.Since(start)

		logger.Info("Server response", zap.String("method", r.Method), zap.String("uri", r.RequestURI), zap.Duration("duration", duration),
			zap.Int("status_code", calcSize.StatusCode),
			zap.Int("response_size", calcSize.Size),
		)
	}
	return http.HandlerFunc(fn)
}

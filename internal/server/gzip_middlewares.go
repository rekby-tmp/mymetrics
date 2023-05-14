package server

import (
	"bytes"
	"compress/gzip"
	"github.com/rekby-tmp/mymetrics/internal/common"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type BufferedResponseWriter struct {
	http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func NewBufferedResponseWriter(w http.ResponseWriter) *BufferedResponseWriter {
	return &BufferedResponseWriter{
		ResponseWriter: w,
		buf:            bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (w *BufferedResponseWriter) Write(content []byte) (int, error) {
	return w.buf.Write(content)
}

func (w *BufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func WithGzipResponse(logger *zap.Logger, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !isAcceptGzipResponse(request) {
			inner.ServeHTTP(writer, request)
			return
		}

		respWriter := NewBufferedResponseWriter(writer)
		inner.ServeHTTP(respWriter, request)
		if needCompress(respWriter.Header().Get("Content-Type")) {
			compressed := bytes.Buffer{}
			compressWriter := gzip.NewWriter(&compressed)
			_, err := io.Copy(compressWriter, &respWriter.buf)
			if err == nil {
				err = compressWriter.Close()
			}
			if err != nil {
				logger.Error("failed to compress response", zap.Error(err))
				http.Error(writer, "failed to compress response", http.StatusInternalServerError)
				return
			}
			respWriter.buf = compressed
			respWriter.Header().Set("Content-Encoding", common.GzipEncoding)
		}

		writer.WriteHeader(respWriter.statusCode)
		_, err := io.Copy(writer, &respWriter.buf)
		if err != nil {
			logger.Warn("failed to write compressed response", zap.Error(err))
		}
	})
}

func WithGzipRequest(logger *zap.Logger, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Encoding") == common.GzipEncoding {
			contentReader, err := gzip.NewReader(request.Body)
			if err != nil {
				logger.Warn("failed to initialize gzip reqder", zap.Error(err))
				http.Error(writer, "failed to read request", http.StatusBadRequest)
				return
			}
			request.Body = contentReader
			request.Header.Del("Content-Encoding")
		}
		inner.ServeHTTP(writer, request)
	})
}

package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestResponseCodes(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		respCode int
	}{
		{
			name:     "counterOk",
			method:   http.MethodPost,
			path:     "/update/counter/test/123",
			respCode: http.StatusOK,
		},
		{
			name:     "counterFloat",
			method:   http.MethodPost,
			path:     "/update/counter/test/123.2",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "counterChars",
			method:   http.MethodPost,
			path:     "/update/counter/test/asd",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "gaugeOK",
			method:   http.MethodPost,
			path:     "/update/gauge/test/123.2",
			respCode: http.StatusOK,
		},
		{
			name:     "gaugeChars",
			method:   http.MethodPost,
			path:     "/update/gauge/test/asd",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "badValueType",
			method:   http.MethodPost,
			path:     "/update/unknown/test/asd",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "getCounter",
			method:   http.MethodGet,
			path:     "/value/counter/test-counter",
			respCode: http.StatusOK,
		},
		{
			name:     "getUnknownCounter",
			method:   http.MethodGet,
			path:     "/value/counter/unknown-counter",
			respCode: http.StatusNotFound,
		},
		{
			name:     "getGauge",
			method:   http.MethodGet,
			path:     "/value/gauge/test-gauge",
			respCode: http.StatusOK,
		},
		{
			name:     "getUnknownGauge",
			method:   http.MethodGet,
			path:     "/value/gauge/unknown-gauge",
			respCode: http.StatusNotFound,
		},
		{
			name:     "getUnknownType",
			method:   http.MethodGet,
			path:     "/value/unknown/unknown",
			respCode: http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMemStorage()
			_ = storage.Store("test-counter", MetricTypeCounter, "1")
			_ = storage.Store("test-gauge", MetricTypeGauge, "2.5")

			server := NewServer("", storage, zaptest.NewLogger(t))

			require.True(t, strings.HasPrefix(test.path, "/"))

			req := httptest.NewRequest(test.method, "http://localhost"+test.path, nil)
			respRec := httptest.NewRecorder()

			server.server.Handler.ServeHTTP(respRec, req)

			resp := respRec.Result()
			require.Equal(t, test.respCode, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

func TestSaveValues(t *testing.T) {
	storage := NewMemStorage()
	s := NewServer("", storage, zaptest.NewLogger(t))

	t.Run("counter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/counter/test1/123", nil)
		s.server.Handler.ServeHTTP(httptest.NewRecorder(), req)
		require.Equal(t, int64(123), storage.counter["test1"])
	})

	t.Run("gauge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/test2/222", nil)
		s.server.Handler.ServeHTTP(httptest.NewRecorder(), req)
		require.Equal(t, float64(222), storage.gauge["test2"])
	})
}

func TestGetValues(t *testing.T) {
	storage := NewMemStorage()
	require.NoError(t, storage.Store("test-counter", MetricTypeCounter, "1"))
	require.NoError(t, storage.Store("test-gauge", MetricTypeGauge, "1.5"))
	s := NewServer("", storage, zaptest.NewLogger(t))

	t.Run("counter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/counter/test-counter", nil)
		resp := httptest.NewRecorder()
		s.server.Handler.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, "1", resp.Body.String())
	})

	t.Run("gauge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/value/gauge/test-gauge", nil)
		resp := httptest.NewRecorder()
		s.server.Handler.ServeHTTP(resp, req)
		require.Equal(t, http.StatusOK, resp.Code)
		require.Equal(t, "1.5", resp.Body.String())
	})
}

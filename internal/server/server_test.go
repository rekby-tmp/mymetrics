package server

import (
	"io"
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
		body     string
		respCode int
	}{
		{
			name:     "counterOk",
			method:   http.MethodPost,
			path:     "/update/counter/test/123",
			body:     "",
			respCode: http.StatusOK,
		},
		{
			name:     "counterFloat",
			method:   http.MethodPost,
			path:     "/update/counter/test/123.2",
			body:     "",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "counterChars",
			method:   http.MethodPost,
			path:     "/update/counter/test/asd",
			body:     "",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "gaugeOK",
			method:   http.MethodPost,
			path:     "/update/gauge/test/123.2",
			body:     "",
			respCode: http.StatusOK,
		},
		{
			name:     "gaugeChars",
			method:   http.MethodPost,
			path:     "/update/gauge/test/asd",
			body:     "",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "badValueType",
			method:   http.MethodPost,
			path:     "/update/unknown/test/asd",
			body:     "",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "getCounter",
			method:   http.MethodGet,
			path:     "/value/counter/test-counter",
			body:     "",
			respCode: http.StatusOK,
		},
		{
			name:     "getUnknownCounter",
			method:   http.MethodGet,
			path:     "/value/counter/unknown-counter",
			body:     "",
			respCode: http.StatusNotFound,
		},
		{
			name:     "getGauge",
			method:   http.MethodGet,
			path:     "/value/gauge/test-gauge",
			body:     "",
			respCode: http.StatusOK,
		},
		{
			name:     "getUnknownGauge",
			method:   http.MethodGet,
			path:     "/value/gauge/unknown-gauge",
			body:     "",
			respCode: http.StatusNotFound,
		},
		{
			name:     "getUnknownType",
			method:   http.MethodGet,
			path:     "/value/unknown/unknown",
			body:     "",
			respCode: http.StatusBadRequest,
		},
		{
			name:     "updateJsonUnknownMetricType",
			method:   http.MethodPost,
			path:     "/update/",
			body:     `{"id": "counter-test", "type": "unknown","delta": 2}`,
			respCode: http.StatusBadRequest,
		},
		{
			name:     "getJsonUnknownMetricType",
			method:   http.MethodPost,
			path:     "/value/",
			body:     `{"id": "counter-test", "type": "unknown"}`,
			respCode: http.StatusBadRequest,
		},
		{
			name:     "getJsonUnknownMetric",
			method:   http.MethodPost,
			path:     "/value/",
			body:     `{"id": "unknown", "type": "counter"}`,
			respCode: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMemStorage()
			_ = storage.Store("test-counter", MetricTypeCounter, int64(1))
			_ = storage.Store("test-gauge", MetricTypeGauge, float64(2.5))

			server := NewServer("", storage, zaptest.NewLogger(t))

			require.True(t, strings.HasPrefix(test.path, "/"))

			req := httptest.NewRequest(test.method, "http://localhost"+test.path, strings.NewReader(test.body))
			if len(test.body) > 0 {
				req.Header.Set("Content-Type", "application/json")
			}
			respRec := httptest.NewRecorder()

			server.server.Handler.ServeHTTP(respRec, req)

			resp := respRec.Result()
			require.Equal(t, test.respCode, resp.StatusCode)
			_ = resp.Body.Close()
		})
	}
}

func TestSaveValuesWithURL(t *testing.T) {
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

func TestGetValuesWithURL(t *testing.T) {
	storage := NewMemStorage()
	require.NoError(t, storage.Store("test-counter", MetricTypeCounter, int64(1)))
	require.NoError(t, storage.Store("test-gauge", MetricTypeGauge, float64(1.5)))
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

func TestGetListMetricts(t *testing.T) {
	e := New(t)
	server := TestServer(e)
	storage := TestStorage(e)
	_ = storage.Store("test-counter-metric", MetricTypeCounter, int64(1))
	_ = storage.Store("test-gauge-metric", MetricTypeGauge, float64(2))

	resp := ServerGetResponse(t, server, "/")
	require.Contains(t, resp, "test-counter-metric")
	require.Contains(t, resp, "test-gauge-metric")
}

func TestUpdateMetricJson(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		e := New(t)
		server := TestServer(e)
		resp, err := http.Post("http://"+server.Endpoint+"/update/", "application/json", strings.NewReader(`{
"id": "counter-test",
"type": "counter",
"delta": 2
}`))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		storage := TestStorage(e)
		res, err := storage.Get("counter-test", MetricTypeCounter)
		require.NoError(t, err)
		require.Equal(t, int64(2), res)
	})
	t.Run("gauge", func(t *testing.T) {
		e := New(t)
		server := TestServer(e)
		resp, err := http.Post("http://"+server.Endpoint+"/update/", "application/json", strings.NewReader(`{
"id": "gauge-test",
"type": "gauge",
"value": 5
}`))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		storage := TestStorage(e)
		res, err := storage.Get("gauge-test", MetricTypeGauge)
		require.NoError(t, err)
		require.Equal(t, float64(5), res)
	})
}

func TestGetJson(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		e := New(t)
		require.NoError(t, TestStorage(e).Store("test-counter", MetricTypeCounter, int64(3)))
		resp, err := http.Post("http://"+TestServer(e).Endpoint+"/value/", "application/json", strings.NewReader(`
{
	"id": "test-counter",
	"type": "counter"
}
`))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		content, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		_ = resp.Body.Close()
		require.JSONEq(t, `
{
	"id": "test-counter",
	"type": "counter",
	"delta": 3
}
`, string(content))
	})
	t.Run("gauge", func(t *testing.T) {
		e := New(t)
		require.NoError(t, TestStorage(e).Store("test-gauge", MetricTypeGauge, float64(5)))
		resp, err := http.Post("http://"+TestServer(e).Endpoint+"/value/", "application/json", strings.NewReader(`
{
	"id": "test-gauge",
	"type": "gauge"
}
`))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		content, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		_ = resp.Body.Close()
		require.JSONEq(t, `
{
	"id": "test-gauge",
	"type": "gauge",
	"value": 5
}
`, string(content))
	})
}

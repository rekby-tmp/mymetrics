package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetricAddress(t *testing.T) {
	agent := &Agent{
		server: "http://asd.com",
	}

	addr := agent.makeURL("test", MetricCounter(15))
	require.Equal(t, "http://asd.com/update/counter/test/15", addr)

	addr = agent.makeURL("eee", MetricGauge(2))
	require.Equal(t, "http://asd.com/update/gauge/eee/2", addr)
}

func TestSend(t *testing.T) {
	var m sync.Mutex
	var values []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		m.Lock()
		defer m.Unlock()

		values = append(values, r.URL.Path)
	}))

	a := &Agent{
		server: server.URL,
	}

	a.values = map[string]Metric{
		"a": MetricCounter(1),
		"b": MetricGauge(2),
	}

	err := a.Send(context.Background())
	require.NoError(t, err)

	sort.Strings(values)
	require.Equal(t, []string{
		"/update/counter/a/1",
		"/update/gauge/b/2",
	}, values)
}

package agent

import (
	"context"
	"encoding/json"
	"github.com/rekby-tmp/mymetrics/internal/common"
	"io"
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
	var values []common.Metrics
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		m.Lock()
		defer m.Unlock()

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var val common.Metrics
		err = json.Unmarshal(body, &val)
		require.NoError(t, err)
		values = append(values, val)
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

	sort.Slice(values, func(i, j int) bool {
		return values[i].ID < values[j].ID
	})
	require.Equal(t, []common.Metrics{
		{
			ID:    "a",
			MType: "counter",
			Delta: &(&[1]int64{1})[0],
			Value: nil,
		},
		{
			ID:    "b",
			MType: "gauge",
			Delta: nil,
			Value: &(&[1]float64{2})[0],
		},
	}, values)
}

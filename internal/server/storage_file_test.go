package server

import (
	"github.com/rekby-tmp/mymetrics/internal/common"
	"github.com/rekby/fixenv/sf"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestFileStorageFlush(t *testing.T) {
	t.Run("force", func(t *testing.T) {
		e := New(t)
		fPath := sf.TempFileNamed(e, "*.json")
		t.Logf("filepath: %q", fPath)
		s := NewFileStorage(fPath, time.Hour*24)
		require.NoError(t, s.Store("test-counter", common.MetricTypeCounter, int64(1)))
		require.NoError(t, s.Store("test-gauge", common.MetricTypeGauge, float64(2)))
		require.NoError(t, s.Flush())

		content, err := os.ReadFile(fPath)
		require.NoError(t, err)
		require.JSONEq(t, `{"Counters":{"test-counter":1},"Gauges":{"test-gauge":2}}`, string(content))
	})
	t.Run("timer", func(t *testing.T) {
		interval := time.Second / 10

		e := New(t)
		fPath := sf.TempFileNamed(e, "*.json")
		t.Logf("filepath: %q", fPath)
		s := NewFileStorage(fPath, interval)
		require.NoError(t, s.Store("test-counter", common.MetricTypeCounter, int64(1)))
		require.NoError(t, s.Store("test-gauge", common.MetricTypeGauge, float64(2)))

		time.Sleep(interval * 2)
		content, err := os.ReadFile(fPath)
		require.NoError(t, err)
		require.JSONEq(t, `{"Counters":{"test-counter":1},"Gauges":{"test-gauge":2}}`, string(content))
		_ = s.Close()
	})
}

func TestFileStorage_LoadFromFile(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		e := New(t)
		fPath := sf.TempFileNamed(e, "*.json")
		require.NoError(t, os.WriteFile(fPath, []byte(`{"Counters":{"test-counter":1},"Gauges":{"test-gauge":2}}`), 0600))

		s := NewFileStorage(fPath, time.Hour*24)
		require.NoError(t, s.LoadFromFile(fPath))

		val, err := s.Get("test-counter", common.MetricTypeCounter)
		require.NoError(t, err)
		require.Equal(t, int64(1), val)

		val, err = s.Get("test-gauge", common.MetricTypeGauge)
		require.NoError(t, err)
		require.Equal(t, float64(2), val)
	})
	t.Run("EmptyJson", func(t *testing.T) {
		e := New(t)
		fPath := sf.TempFileNamed(e, "*.json")
		require.NoError(t, os.WriteFile(fPath, []byte(`{}`), 0600))

		s := NewFileStorage(fPath, time.Hour*24)
		require.NoError(t, s.LoadFromFile(fPath))
		require.NotNil(t, s.storage.counter)
		require.NotNil(t, s.storage.gauge)

		_, err := s.Get("test-unknown", common.MetricTypeCounter)
		require.ErrorIs(t, err, ErrNotFound)
	})
}

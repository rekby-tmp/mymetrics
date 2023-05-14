package server

import (
	"fmt"
	"github.com/rekby-tmp/mymetrics/internal/common"
	"reflect"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

func (m *MemStorage) Get(name string, metricType common.MetricType) (value any, err error) {
	var val any
	var ok bool
	switch metricType {
	case common.MetricTypeCounter:
		val, ok = m.counter[name]
	case common.MetricTypeGauge:
		val, ok = m.gauge[name]
	default:
		return nil, fmt.Errorf("failed to get metric type %q/%q: %w", metricType, name, ErrUnknownMetricType)
	}

	if !ok {
		return "", fmt.Errorf("failed to get metric value %q/%q: %w", metricType, name, ErrNotFound)
	}
	return val, nil
}

func (m *MemStorage) List() (map[common.MetricType][]string, error) {
	counters := make([]string, len(m.counter))
	for name := range m.counter {
		counters = append(counters, name)
	}

	gauges := make([]string, len(m.gauge))
	for name := range m.gauge {
		gauges = append(gauges, name)
	}

	return map[common.MetricType][]string{
		common.MetricTypeCounter: counters,
		common.MetricTypeGauge:   gauges,
	}, nil
}

func (m *MemStorage) Store(name string, metricType common.MetricType, value any) error {
	switch metricType {
	case common.MetricTypeCounter:
		val, ok := value.(int64)
		if !ok {
			return fmt.Errorf("bad value type. Need int64, has: %q", reflect.TypeOf(value).String())
		}
		m.counter[name] += val
	case common.MetricTypeGauge:
		val, ok := value.(float64)
		if !ok {
			return fmt.Errorf("bad value type. Need float64, has: %q", reflect.TypeOf(value).String())
		}
		m.gauge[name] = val
	default:
		return fmt.Errorf("failed to store %q/%q: %w", name, metricType, ErrUnknownMetricType)
	}

	return nil
}

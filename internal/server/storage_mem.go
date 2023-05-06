package server

import (
	"fmt"
	"strconv"
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

func (m *MemStorage) Get(name string, metricType MetricType) (value string, err error) {
	var val any
	var ok bool
	switch metricType {
	case MetricTypeCounter:
		val, ok = m.counter[name]
	case MetricTypeGauge:
		val, ok = m.gauge[name]
	default:
		return "", fmt.Errorf("failed to get metric type %q/%q: %w", metricType, name, errUnknownMetricType)
	}

	if !ok {
		return "", fmt.Errorf("failed to get metric value %q/%q: %w", metricType, name, errNotFound)
	}
	return fmt.Sprint(val), nil
}

func (m *MemStorage) List() (map[MetricType][]string, error) {
	counters := make([]string, len(m.counter))
	for name := range m.counter {
		counters = append(counters, name)
	}

	gauges := make([]string, len(m.gauge))
	for name := range m.gauge {
		gauges = append(gauges, name)
	}

	return map[MetricType][]string{
		MetricTypeCounter: counters,
		MetricTypeGauge:   gauges,
	}, nil
}

func (m *MemStorage) Store(name string, metricType MetricType, value string) error {
	switch metricType {
	case MetricTypeCounter:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse counter value %q %q: %w", name, value, errBadValue)
		}
		m.counter[name] += val
	case MetricTypeGauge:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse gauge value %q %q: %w", name, value, errBadValue)
		}
		m.gauge[name] = val
	default:
		return fmt.Errorf("failed to store %q/%q: %w", name, metricType, errUnknownMetricType)
	}

	return nil
}

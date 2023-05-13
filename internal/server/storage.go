package server

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrNotFound          = errors.New("value not found")
	ErrBadValue          = errors.New("failed value parse")
	ErrUnknownMetricType = errors.New("unknown metric type")
)

type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

type Storage interface {
	Get(name string, metricType MetricType) (val any, err error)
	List() (map[MetricType][]string, error)

	// Store - save value.
	// val - int64 for counter value and float64 for gauge
	Store(name string, metricType MetricType, val any) (err error)
}

func ParseMetricValue(m MetricType, val string) (res any, err error) {
	switch m {
	case MetricTypeCounter:
		return strconv.ParseInt(val, 10, 64)
	case MetricTypeGauge:
		return strconv.ParseFloat(val, 64)
	default:
		return nil, fmt.Errorf("receive metric type for parse metric value %q: %w", m, ErrUnknownMetricType)
	}
}

package server

import "errors"

var (
	errNotFound          = errors.New("value not found")
	errBadValue          = errors.New("failed value parse")
	errUnknownMetricType = errors.New("unknown metric type")
)

type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

type Storage interface {
	Get(name string, metricType MetricType) (val string, err error)
	List() (map[MetricType][]string, error)
	Store(name string, metricType MetricType, val string) (err error)
}

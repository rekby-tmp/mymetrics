package common

type MetricType string

func (mt MetricType) String() string {
	return string(mt)
}

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

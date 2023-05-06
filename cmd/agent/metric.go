package main

import (
	"strconv"
)

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

type Metric interface {
	String() string
	Type() MetricType
}

type MetricCounter int64

func (c MetricCounter) Type() MetricType {
	return MetricTypeCounter
}

func (c MetricCounter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

type MetricGauge float64

func (g MetricGauge) Type() MetricType {
	return MetricTypeGauge
}

func (g MetricGauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

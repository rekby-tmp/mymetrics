package agent

import (
	"github.com/rekby-tmp/mymetrics/internal/common"
	"strconv"
)

type Metric interface {
	String() string
	Type() common.MetricType
	Value() any
}

type MetricCounter int64

func (c MetricCounter) Type() common.MetricType {
	return common.MetricTypeCounter
}

func (c MetricCounter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func (c MetricCounter) Value() any {
	return int64(c)
}

type MetricGauge float64

func (g MetricGauge) Type() common.MetricType {
	return common.MetricTypeGauge
}

func (g MetricGauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func (g MetricGauge) Value() any {
	return float64(g)
}

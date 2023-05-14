package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rekby-tmp/mymetrics/internal/common"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type MetricType string

func (mt MetricType) String() string {
	return string(mt)
}

type Agent struct {
	server       string
	pollInterval time.Duration
	sendInterval time.Duration

	m      sync.Mutex
	polls  int
	values map[string]Metric
}

func NewAgent(server string, pollInterval, sendInterval time.Duration) *Agent {
	return &Agent{
		server:       server,
		pollInterval: pollInterval,
		sendInterval: sendInterval,

		values: map[string]Metric{},
	}
}

func (a *Agent) Start() {
	go func() {
		ticker := time.NewTicker(a.pollInterval)
		a.Poll()
		for {
			<-ticker.C
			a.Poll()
		}
	}()

	go func() {
		ticker := time.NewTicker(a.sendInterval)
		for {
			<-ticker.C
			sendCtx, cancel := context.WithTimeout(context.Background(), a.sendInterval)
			err := a.Send(sendCtx)
			if err != nil {
				log.Printf("Error while send metric: %+v", err)
			}
			cancel()
		}
	}()

	var stopChan chan bool
	<-stopChan
}

func (a *Agent) Poll() {
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)

	a.m.Lock()
	defer a.m.Unlock()

	a.polls++
	a.values["PollCount"] = MetricCounter(a.polls)
	a.values["RandomValue"] = MetricGauge(rand.Float64())

	a.values["Alloc"] = MetricGauge(memStat.Alloc)
	a.values["BuckHashSys"] = MetricGauge(memStat.BuckHashSys)
	a.values["Frees"] = MetricGauge(memStat.Frees)
	a.values["GCCPUFraction"] = MetricGauge(memStat.GCCPUFraction)
	a.values["GCSys"] = MetricGauge(memStat.GCSys)
	a.values["HeapAlloc"] = MetricGauge(memStat.HeapAlloc)
	a.values["HeapIdle"] = MetricGauge(memStat.HeapIdle)
	a.values["HeapInuse"] = MetricGauge(memStat.HeapInuse)
	a.values["HeapObjects"] = MetricGauge(memStat.HeapObjects)
	a.values["HeapReleased"] = MetricGauge(memStat.HeapReleased)
	a.values["HeapSys"] = MetricGauge(memStat.HeapSys)
	a.values["LastGC"] = MetricGauge(memStat.LastGC)
	a.values["Lookups"] = MetricGauge(memStat.Lookups)
	a.values["MCacheInuse"] = MetricGauge(memStat.MCacheInuse)
	a.values["MCacheSys"] = MetricGauge(memStat.MCacheSys)
	a.values["MSpanInuse"] = MetricGauge(memStat.MSpanInuse)
	a.values["MSpanSys"] = MetricGauge(memStat.MSpanSys)
	a.values["Mallocs"] = MetricGauge(memStat.Mallocs)
	a.values["NextGC"] = MetricGauge(memStat.NextGC)
	a.values["NumForcedGC"] = MetricGauge(memStat.NumForcedGC)
	a.values["NumGC"] = MetricGauge(memStat.NumGC)
	a.values["OtherSys"] = MetricGauge(memStat.OtherSys)
	a.values["PauseTotalNs"] = MetricGauge(memStat.PauseTotalNs)
	a.values["StackInuse"] = MetricGauge(memStat.StackInuse)
	a.values["StackSys"] = MetricGauge(memStat.StackSys)
	a.values["Sys"] = MetricGauge(memStat.Sys)
	a.values["TotalAlloc"] = MetricGauge(memStat.TotalAlloc)
}

func (a *Agent) Send(ctx context.Context) error {
	localValues := a.cloneValues()

	for k, v := range localValues {
		err := a.sendValue(ctx, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) sendValue(ctx context.Context, name string, value Metric) error {
	val := common.Metrics{
		ID:    name,
		MType: value.Type().String(),
	}

	switch value.Type() {
	case MetricTypeCounter:
		v := value.Value().(int64)
		val.Delta = &v
	case MetricTypeGauge:
		v := value.Value().(float64)
		val.Value = &v
	default:
		return fmt.Errorf("unknown metric type: %q", value.Type().String())
	}

	content, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed marshal value: %w", err)
	}

	address := a.server + "/update/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed create http request for send metric: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status: %v", res.Status)
	}
	return nil
}

func (a *Agent) makeURL(name string, value Metric) string {
	return fmt.Sprintf("%v/update/%v/%v/%v", a.server, value.Type(), name, value.String())
}

func (a *Agent) cloneValues() map[string]Metric {
	a.m.Lock()
	defer a.m.Unlock()

	res := make(map[string]Metric, len(a.values))

	for k, v := range a.values {
		res[k] = v
	}

	return res
}

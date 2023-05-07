package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/rekby-tmp/mymetrics/internal/agent"
)

type Config struct {
	Endpoint       string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "endpoint")
	pollIntervalSeconds := flag.Int("p", 2, "Poll interval (seconds)")
	reportIntervalSeconds := flag.Int("r", 10, "Report interval")
	flag.Parse()

	cfg.PollInterval = time.Duration(*pollIntervalSeconds) * time.Second
	cfg.ReportInterval = time.Duration(*reportIntervalSeconds) * time.Second

	if !strings.HasPrefix(cfg.Endpoint, "http://") {
		cfg.Endpoint = "http://" + cfg.Endpoint
	}

	log.Printf("Start agent with config: %#v\n", cfg)

	agent := agent.NewAgent(cfg.Endpoint, cfg.PollInterval, cfg.ReportInterval)
	agent.Start()
}

package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rekby-tmp/mymetrics/internal/agent"
)

type Config struct {
	Endpoint              string `env:"ADDRESS"`
	ReportIntervalSeconds int    `env:"REPORT_INTERVAL"`
	PollIntervalSeconds   int    `env:"POLL_INTERVAL"`
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "endpoint")
	flag.IntVar(&cfg.PollIntervalSeconds, "p", 2, "Poll interval (seconds)")
	flag.IntVar(&cfg.ReportIntervalSeconds, "r", 10, "Report interval")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed parse env variables: %v", err)
	}

	if !strings.HasPrefix(cfg.Endpoint, "http://") {
		cfg.Endpoint = "http://" + cfg.Endpoint
	}

	log.Printf("Start agent with config: %#v\n", cfg)

	agent := agent.NewAgent(cfg.Endpoint, time.Duration(cfg.PollIntervalSeconds)*time.Second, time.Duration(cfg.ReportIntervalSeconds)*time.Second)
	agent.Start()
}

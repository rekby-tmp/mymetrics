package main

import (
	"errors"
	"flag"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/rekby-tmp/mymetrics/internal/server"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint string `env:"ADDRESS"`
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed initialize logger: %v", err)
	}

	var cfg Config
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "Endpoint")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed parse env variables: %v", err)
	}

	logger.Info("Start server", zap.Reflect("config", cfg))
	s := server.NewServer(cfg.Endpoint, server.NewMemStorage(), logger)
	err = s.Start()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

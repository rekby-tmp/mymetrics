package main

import (
	"errors"
	"flag"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/rekby-tmp/mymetrics/internal/server"
)

type Config struct {
	Endpoint string `env:"ADDRESS"`
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "Endpoint")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed parse env variables: %v", err)
	}

	log.Printf("Start server with config: %#v\n", cfg)
	s := server.NewServer(cfg.Endpoint, server.NewMemStorage())
	err := s.Start()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env"
	_ "github.com/lib/pq"
	"github.com/rekby-tmp/mymetrics/internal/server"
	"go.uber.org/zap"
)

type Config struct {
	Endpoint           string `env:"ADDRESS"`
	StoreInterval      int    `env:"STORE_INTERVAL"`
	StorePath          string `env:"FILE_STORAGE_PATH"`
	Restore            bool   `env:"RESTORE"`
	DBConnectionString string `env:"DATABASE_DSN"`
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed initialize logger: %v", err)
	}

	var cfg Config
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "Endpoint")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Store interval, seconds")
	flag.StringVar(&cfg.StorePath, "f", "/tmp/metrics-db.json", "Storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore state from storage path")
	flag.StringVar(&cfg.DBConnectionString, "d", "", "Database connection string")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed parse env variables: %v", err)
	}

	logger.Info("Init storage")
	storage := server.NewFileStorage(cfg.StorePath, time.Duration(cfg.StoreInterval)*time.Second)
	if cfg.Restore {
		if _, err := os.Stat(cfg.StorePath); !os.IsNotExist(err) {
			err = storage.LoadFromFile(cfg.StorePath)
			if err != nil {
				log.Fatalf("failed load old state from file %q: %+v", cfg.StorePath, err)
			}
		}
	}

	defer func() {
		err := storage.Flush()
		if err != nil {
			log.Fatalf("failed to flush storage on exit to file %q: %+v", cfg.StorePath, err)
		}
		_ = storage.Close()
	}()

	logger.Info("Start server", zap.Reflect("config", cfg))
	s := server.NewServer(cfg.Endpoint, storage, logger)
	s.DBConnectionString = cfg.DBConnectionString
	err = s.Start()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

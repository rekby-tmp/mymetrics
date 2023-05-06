package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/rekby-tmp/mymetrics/internal/server"
)

func main() {
	s := server.NewServer("localhost:8080", server.NewMemStorage())
	err := s.Start()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

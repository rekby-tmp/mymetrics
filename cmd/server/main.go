package main

import (
	"errors"
	"log"
	"net/http"
)

func main() {
	server := NewServer("localhost:8080", NewMemStorage())
	err := server.Start()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

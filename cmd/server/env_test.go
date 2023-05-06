package main

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/rekby/fixenv"
)

func NewEnv(t *testing.T) fixenv.Env {
	return fixenv.NewEnv(t)
}

func MetricsServer(e fixenv.Env) *Server {
	return e.CacheWithCleanup(nil, nil, func() (res interface{}, cleanup fixenv.FixtureCleanupFunc, err error) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, nil, err
		}
		addr := listener.Addr().String()
		err = listener.Close()
		if err != nil {
			return nil, nil, err
		}

		server := NewServer(addr, NewMemStorage())
		var startErr error
		go func() {
			startErr = server.Start()
		}()

		for {
			if startErr != nil {
				return nil, nil, err
			}
			resp, err := http.Get("http://" + server.endpoint)
			if err == nil {
				_ = resp.Body.Close()

				return server, func() { _ = server.Shutdown(context.Background()) }, nil
			}
		}
	}).(*Server)
}

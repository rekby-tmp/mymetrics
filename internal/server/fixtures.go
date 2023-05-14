package server

import (
	"context"
	"github.com/rekby/fixenv"
	"github.com/rekby/fixenv/sf"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type Env = fixenv.Env

func New(t testing.TB) Env {
	return fixenv.New(t)
}

func TestLogger(e Env) *zap.Logger {
	return fixenv.Cache(e, nil, nil, func() (*zap.Logger, error) {
		return zaptest.NewLogger(e.T().(zaptest.TestingT)), nil
	})
}

func TestStorage(e Env) Storage {
	return fixenv.Cache(e, nil, nil, func() (Storage, error) {
		return NewMemStorage(), nil
	})
}

func TestServer(e Env) *Server {
	return fixenv.CacheWithCleanup(e, nil, nil, func() (*Server, fixenv.FixtureCleanupFunc, error) {
		server := NewServer(sf.FreeLocalTCPAddress(e), TestStorage(e), TestLogger(e))
		go func() {
			server.Start()
		}()
		time.Sleep(time.Millisecond * 10)
		clean := func() { _ = server.Shutdown(context.Background()) }
		return server, clean, nil
	})
}

func ServerGetResponse(t testing.TB, s *Server, path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	resp, err := http.Get("http://" + s.Endpoint + path)
	if err != nil {
		t.Fatalf("failed get response for path %q: %+v", path, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed read content for path %q: %+v", path, err)
	}

	return string(body)
}

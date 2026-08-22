package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

type databaseStub struct{ pingErr error }

func (d databaseStub) PingContext(context.Context) error { return d.pingErr }
func (d databaseStub) QueryTimeout() time.Duration       { return time.Second }

func TestHealthEndpointsHaveSeparateSemantics(t *testing.T) {
	router := NewRouter(databaseStub{pingErr: errors.New("database unavailable")}, time.Second, zap.NewNop())

	live := httptest.NewRecorder()
	router.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("liveness status = %d", live.Code)
	}

	ready := httptest.NewRecorder()
	router.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness status = %d", ready.Code)
	}
	if ready.Header().Get("X-Request-ID") == "" {
		t.Fatal("readiness response has no request ID")
	}
}

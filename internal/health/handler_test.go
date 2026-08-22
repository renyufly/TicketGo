package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type stubPinger struct{ err error }

func (p stubPinger) PingContext(context.Context) error { return p.err }

func TestReadyReportsDependencyFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health/ready", New(stubPinger{err: errors.New("offline")}, time.Second, zap.NewNop()).Ready)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "offline") {
		t.Fatal("response leaked internal database error")
	}
	if !strings.Contains(recorder.Body.String(), "dependency_unavailable") {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

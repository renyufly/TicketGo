package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDPreservesCallerValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "known-request")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if got := recorder.Header().Get(RequestIDHeader); got != "known-request" {
		t.Fatalf("X-Request-ID = %q", got)
	}
}

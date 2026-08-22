package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ticketgo/internal/health"
	mw "ticketgo/internal/httpapi/middleware"
	"ticketgo/pkg/response"
)

type DatabaseHealth interface {
	PingContext(context.Context) error
	QueryTimeout() time.Duration
}

func NewRouter(db DatabaseHealth, requestTimeout time.Duration, logger *zap.Logger) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(mw.RequestID(), mw.AccessLog(logger), mw.Recovery(logger), mw.Timeout(requestTimeout))

	healthHandler := health.New(db, db.QueryTimeout(), logger)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)
	router.NoRoute(func(c *gin.Context) {
		response.Error(c, response.NewError(http.StatusNotFound, "not_found", "route not found", nil))
	})
	router.NoMethod(func(c *gin.Context) {
		response.Error(c, response.NewError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil))
	})
	return router
}

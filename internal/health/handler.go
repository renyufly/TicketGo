package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ticketgo/pkg/response"
)

type Pinger interface{ PingContext(context.Context) error }

type Handler struct {
	db      Pinger
	timeout time.Duration
	logger  *zap.Logger
}

func New(db Pinger, timeout time.Duration, logger *zap.Logger) Handler {
	return Handler{db: db, timeout: timeout, logger: logger}
}

func (h Handler) Live(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Warn("readiness dependency check failed", zap.String("dependency", "postgresql"), zap.Error(err))
		response.Error(c, response.NewError(http.StatusServiceUnavailable, "dependency_unavailable", "PostgreSQL is not ready", err))
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "ok", "dependencies": gin.H{"postgresql": "ok"}})
}

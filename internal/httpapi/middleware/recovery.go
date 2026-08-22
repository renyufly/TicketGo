package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ticketgo/pkg/response"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered", zap.Any("panic", recovered), zap.String("request_id", response.RequestID(c)), zap.Stack("stack"))
		response.Error(c, response.NewError(http.StatusInternalServerError, "internal_error", "internal server error", nil))
		c.Abort()
	})
}

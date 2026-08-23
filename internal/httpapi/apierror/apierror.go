package apierror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"ticketgo/internal/domain"
	"ticketgo/pkg/response"
)

func WriteValidation(c *gin.Context, cause error) {
	response.Error(c, response.NewError(http.StatusBadRequest, "validation_error", "request validation failed", cause))
}
func Write(c *gin.Context, err error) {
	_ = c.Error(err)
	status, code, message := http.StatusInternalServerError, "internal_error", "internal server error"
	var de *domain.Error
	if errors.As(err, &de) {
		message = de.Message
	}
	switch {
	case errors.Is(err, domain.ErrInvalid):
		status, code = http.StatusBadRequest, "invalid_request"
	case errors.Is(err, domain.ErrUnauthenticated):
		status, code = http.StatusUnauthorized, "unauthenticated"
	case errors.Is(err, domain.ErrForbidden):
		status, code = http.StatusForbidden, "forbidden"
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, domain.ErrOutOfStock):
		status, code = http.StatusConflict, "out_of_stock"
	case errors.Is(err, domain.ErrActivityClosed):
		status, code = http.StatusConflict, "activity_unavailable"
	case errors.Is(err, domain.ErrBusy):
		status, code = http.StatusServiceUnavailable, "concurrency_busy"
	}
	response.Error(c, response.NewError(status, code, message, err))
}

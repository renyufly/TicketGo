package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

type AppError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Cause }

func NewError(status int, code, message string, cause error) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Cause: cause}
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "request_id": RequestID(c)})
}

func Error(c *gin.Context, err error) {
	appErr := &AppError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "internal server error"}
	if !errors.As(err, &appErr) {
		appErr = &AppError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "internal server error", Cause: err}
	}
	c.JSON(appErr.Status, gin.H{"code": appErr.Code, "message": appErr.Message, "request_id": RequestID(c)})
}

func RequestID(c *gin.Context) string {
	value, _ := c.Get(RequestIDKey)
	requestID, _ := value.(string)
	return requestID
}

package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"ticketgo/pkg/response"
)

const claimsKey = "auth_claims"

func Required(manager *Manager, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, response.NewError(http.StatusUnauthorized, "unauthenticated", "valid bearer token required", nil))
			c.Abort()
			return
		}
		claims, err := manager.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			response.Error(c, response.NewError(http.StatusUnauthorized, "unauthenticated", "valid bearer token required", err))
			c.Abort()
			return
		}
		var role, status string
		if err := db.QueryRowContext(c.Request.Context(), `SELECT role,status FROM users WHERE id=$1`, claims.UserID).Scan(&role, &status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.Error(c, response.NewError(http.StatusUnauthorized, "unauthenticated", "valid bearer token required", err))
			} else {
				response.Error(c, response.NewError(http.StatusInternalServerError, "internal_error", "internal server error", err))
			}
			c.Abort()
			return
		}
		if status != "active" {
			response.Error(c, response.NewError(http.StatusForbidden, "user_disabled", "user account is disabled", nil))
			c.Abort()
			return
		}
		claims.Role = role
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsFrom(c)
		if !ok || claims.Role != "admin" {
			response.Error(c, response.NewError(http.StatusForbidden, "forbidden", "administrator role required", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

func ClaimsFrom(c *gin.Context) (Claims, bool) {
	value, ok := c.Get(claimsKey)
	if !ok {
		return Claims{}, false
	}
	claims, ok := value.(Claims)
	return claims, ok
}

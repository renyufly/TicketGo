// Gin 中的身份认证 + 权限控制中间件
// JWT + Gin Middleware + RBAC（角色权限控制） 结构
// Required()    → 你登录了吗？
// AdminOnly()   → 你是管理员吗？
// ClaimsFrom()  → 当前登录的是谁？

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

// 表示：登录用户才能访问
func Required(manager *Manager, db *sql.DB) gin.HandlerFunc {
	// 认证中间件Middleware

	return func(c *gin.Context) {
		// 读取 Authorization -> 是否为 Bearer xxx？
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, response.NewError(http.StatusUnauthorized, "unauthenticated", "valid bearer token required", nil))
			c.Abort()
			return
		}

		// 解析 JWT， 拿到 UserID
		claims, err := manager.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			response.Error(c, response.NewError(http.StatusUnauthorized, "unauthenticated", "valid bearer token required", err))
			c.Abort()
			return
		}

		// 查询数据库中的 role、status
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

		// 把认证后的用户信息存进当前请求的 gin.Context，然后放行
		claims.Role = role
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// 必须是管理员：验证是否有 admin 权限
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Context 中取出刚才 Required() 保存的 Claims
		claims, ok := ClaimsFrom(c)
		if !ok || claims.Role != "admin" {
			// 不是管理员就返回：403 Forbidden
			response.Error(c, response.NewError(http.StatusForbidden, "forbidden", "administrator role required", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

// 获取当前用户Claims：知道当前登录的是谁
func ClaimsFrom(c *gin.Context) (Claims, bool) {
	value, ok := c.Get(claimsKey)
	if !ok {
		return Claims{}, false
	}
	// claims.UserID 就是当前登录用户 ID
	claims, ok := value.(Claims)
	return claims, ok
}

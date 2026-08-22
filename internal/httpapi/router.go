// 整个后端的 路由总装配中心
// 把中间件、认证、业务模块和 API 地址全部组装起来
/*
创建 Gin HTTP 服务器的路由结构，
把中间件、JWT 认证、管理员权限，以及 user/item/activity/order 
等业务 Handler 统一组装并映射到具体 API.
本身基本不负责业务逻辑，而是负责：
哪个 URL → 经过哪些中间件 → 调哪个 Handler
*/

package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ticketgo/internal/activity"
	"ticketgo/internal/auth"
	"ticketgo/internal/config"
	"ticketgo/internal/health"
	mw "ticketgo/internal/httpapi/middleware"
	"ticketgo/internal/inventory"
	"ticketgo/internal/item"
	"ticketgo/internal/order"
	"ticketgo/internal/user"
	"ticketgo/pkg/response"
)

// 定义 Router 所需要的数据库能力，例如检查数据库是否正常、获取 SQL 连接
type DatabaseHealth interface {
	PingContext(context.Context) error
	QueryTimeout() time.Duration
	SQL() *sql.DB
}

// 创建 Gin Router - HTTP 路由器
func NewRouter(db DatabaseHealth, cfg config.Config, logger *zap.Logger) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.HandleMethodNotAllowed = true

	// 注册全局中间件: 每个请求都会经过：
	// RequestID：给请求生成唯一 ID
	// AccessLog：记录访问日志
	// Recovery：防止 panic 导致服务器崩溃
	// Timeout：限制请求最大执行时间
	router.Use(mw.RequestID(), mw.AccessLog(logger), mw.Recovery(logger), mw.Timeout(cfg.HTTP.RequestTimeout))

	// 健康检查接口
	healthHandler := health.New(db, db.QueryTimeout(), logger)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	tokens := auth.NewManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)
	
	// 初始化各业务模块
	// Repository(数据库访问) → Service(业务逻辑) → Handler(HTTP处理)
	userHandler := user.NewHandler(user.NewService(user.NewRepository(db.SQL()), tokens, cfg.Auth.AllowAdminSelfRegistration))
	itemHandler := item.NewHandler(item.NewService(item.NewRepository(db.SQL())))
	activityHandler := activity.NewHandler(activity.NewService(activity.NewRepository(db.SQL())))
	orderHandler := order.NewHandler(order.NewService(db.SQL(), order.NewRepository(db.SQL()), inventory.NewRepository()))

	// 注册公开接口
	v1 := router.Group("/api/v1")
	v1.POST("/users", userHandler.Create)
	v1.POST("/login", userHandler.Login)

	// 注册需要登录的接口:
	// 之后注册到 authorized 的接口都会先检查 JWT
	authorized := v1.Group("")
	authorized.Use(auth.Required(tokens, db.SQL()))

	// GET /api/v1/orders -> auth.Required()
	// -> 验证 JWT -> orderHandler.List()
	authorized.GET("/users/me", userHandler.Me)
	authorized.GET("/items", itemHandler.List)
	authorized.GET("/items/:id", itemHandler.Get)
	authorized.GET("/activities", activityHandler.List)
	authorized.GET("/activities/:id", activityHandler.Get)
	authorized.POST("/activities/:id/seckill", orderHandler.Seckill)
	authorized.GET("/orders", orderHandler.List)
	authorized.GET("/orders/:id", orderHandler.Get)
	authorized.POST("/orders/:id/cancel", orderHandler.Cancel)
	
	// 管理员接口:
	// 在“必须登录”的基础上，再增加：必须是 admin
	admin := authorized.Group("")
	admin.Use(auth.AdminOnly())

	admin.POST("/items", itemHandler.Create)
	admin.POST("/activities", activityHandler.Create)

	// 处理不存在的 URL
	router.NoRoute(func(c *gin.Context) {
		response.Error(c, response.NewError(http.StatusNotFound, "not_found", "route not found", nil))
	})

	// 处理 URL 存在但 HTTP 方法不对 (如Get变Post)
	router.NoMethod(func(c *gin.Context) {
		response.Error(c, response.NewError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil))
	})

	return router
}

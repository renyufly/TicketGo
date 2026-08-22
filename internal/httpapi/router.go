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

type DatabaseHealth interface {
	PingContext(context.Context) error
	QueryTimeout() time.Duration
	SQL() *sql.DB
}

func NewRouter(db DatabaseHealth, cfg config.Config, logger *zap.Logger) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(mw.RequestID(), mw.AccessLog(logger), mw.Recovery(logger), mw.Timeout(cfg.HTTP.RequestTimeout))

	healthHandler := health.New(db, db.QueryTimeout(), logger)
	router.GET("/health/live", healthHandler.Live)
	router.GET("/health/ready", healthHandler.Ready)

	tokens := auth.NewManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL)
	userHandler := user.NewHandler(user.NewService(user.NewRepository(db.SQL()), tokens, cfg.Auth.AllowAdminSelfRegistration))
	itemHandler := item.NewHandler(item.NewService(item.NewRepository(db.SQL())))
	activityHandler := activity.NewHandler(activity.NewService(activity.NewRepository(db.SQL())))
	orderHandler := order.NewHandler(order.NewService(db.SQL(), order.NewRepository(db.SQL()), inventory.NewRepository()))

	v1 := router.Group("/api/v1")
	v1.POST("/users", userHandler.Create)
	v1.POST("/login", userHandler.Login)
	authorized := v1.Group("")
	authorized.Use(auth.Required(tokens, db.SQL()))
	authorized.GET("/users/me", userHandler.Me)
	authorized.GET("/items", itemHandler.List)
	authorized.GET("/items/:id", itemHandler.Get)
	authorized.GET("/activities", activityHandler.List)
	authorized.GET("/activities/:id", activityHandler.Get)
	authorized.POST("/activities/:id/seckill", orderHandler.Seckill)
	authorized.GET("/orders", orderHandler.List)
	authorized.GET("/orders/:id", orderHandler.Get)
	authorized.POST("/orders/:id/cancel", orderHandler.Cancel)
	admin := authorized.Group("")
	admin.Use(auth.AdminOnly())
	admin.POST("/items", itemHandler.Create)
	admin.POST("/activities", activityHandler.Create)
	router.NoRoute(func(c *gin.Context) {
		response.Error(c, response.NewError(http.StatusNotFound, "not_found", "route not found", nil))
	})
	router.NoMethod(func(c *gin.Context) {
		response.Error(c, response.NewError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil))
	})
	return router
}

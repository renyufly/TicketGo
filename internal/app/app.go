// 整个 Go 后端应用的启动和关闭入口
// 初始化资源 → 启动服务 → 监听退出 → Graceful Shutdown → 释放资源

package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"ticketgo/internal/config"
	"ticketgo/internal/httpapi"
	"ticketgo/pkg/database"
)

func Run(ctx context.Context, cfg config.Config, logger *zap.Logger) error {
	// 	接收三个核心依赖：
	// ctx：控制程序生命周期，例如收到 Ctrl+C / SIGTERM 后通知关闭。
	// cfg：配置，包括数据库地址、HTTP 端口、超时时间等。
	// logger：Zap 日志记录器

	// 连接 PostgreSQL
	db, err := database.Open(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer func() {
		// 确保 Run 结束时数据库连接池被关闭
		if err := db.Close(); err != nil {
			logger.Error("close PostgreSQL connection pool", zap.Error(err))
		}
	}()

	// 创建 Go 标准库的 HTTP 服务器.
	// Handler 是真正处理 HTTP 请求的 Router
	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           httpapi.NewRouter(db, cfg, logger),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	serverErr := make(chan error, 1)

	// 用 goroutine 启动 HTTP Server，否则 ListenAndServe() 会一直阻
	go func() {
		logger.Info("HTTP server started", zap.String("address", cfg.HTTP.Address))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		// 收到关闭信号
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		// HTTP Server 自己停止/报错
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}

	// 执行 Graceful Shutdown（优雅关闭)
	// 停止接受新请求，但给正在处理的请求一定时间完成，而不是直接把服务器杀掉
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful HTTP shutdown: %w", err)
	}

	logger.Info("HTTP server stopped cleanly")
	return nil
}

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
	db, err := database.Open(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("close PostgreSQL connection pool", zap.Error(err))
		}
	}()

	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           httpapi.NewRouter(db, cfg.HTTP.RequestTimeout, logger),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", zap.String("address", cfg.HTTP.Address))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful HTTP shutdown: %w", err)
	}

	logger.Info("HTTP server stopped cleanly")
	return nil
}

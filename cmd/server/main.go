package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/srivickynesh/light-simulator/internal/config"
	"github.com/srivickynesh/light-simulator/internal/handlers"
	"github.com/srivickynesh/light-simulator/internal/middleware"
	"github.com/srivickynesh/light-simulator/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logLevel := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	staticSub, err := fs.Sub(web.Content, "static")
	if err != nil {
		return fmt.Errorf("embedded static fs: %w", err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	templateSub, err := fs.Sub(web.Content, "templates")
	if err != nil {
		return fmt.Errorf("embedded template fs: %w", err)
	}
	pages := handlers.NewPagesFS(templateSub, !cfg.IsProd(), logger)
	pages.RegisterRoutes(mux)

	// Register API routes
	api := handlers.NewAPI()
	api.RegisterRoutes(mux)

	// Register upload routes
	upload := handlers.NewUpload(cfg.UploadDir, cfg.MaxUploadMB, logger)
	upload.RegisterRoutes(mux)

	// Apply middleware stack
	handler := middleware.Chain(
		mux,
		middleware.Recover(logger),
		middleware.Logger(logger),
		middleware.CORS,
		middleware.SecurityHeaders,
	)

	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", cfg.Addr(), "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("shutdown signal received", "signal", sig)
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

// Command api is the entrypoint for the REST API server.
// It wires together configuration, the data store, the HTTP router,
// and the server lifecycle (start + graceful shutdown).
//
// Run with:
//
//	go run ./cmd/api
//
// or build and run:
//
//	go build -o bin/api ./cmd/api && ./bin/api
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/go-api-template/internal/api"
	"github.com/user/go-api-template/internal/config"
	"github.com/user/go-api-template/internal/store"
)

func main() {
	// ── Structured JSON logger ──────────────────────────────────────────────
	// slog.NewJSONHandler writes machine-readable JSON to stdout.
	// Replace with slog.NewTextHandler for human-readable local dev output.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── Configuration ────────────────────────────────────────────────────────
	cfg := config.Load()
	slog.Info("config loaded", "addr", cfg.Addr(), "trusted_proxy", cfg.TrustedProxy)

	// ── Data store ──────────────────────────────────────────────────────────
	// Swap NewMemoryStore() for a database-backed implementation here.
	// The rest of the application only sees the store.TodoStore interface.
	s := store.NewMemoryStore()

	// ── HTTP server ─────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:    cfg.Addr(),
		Handler: api.NewRouter(cfg, s),

		// Always set explicit timeouts. The zero value means no timeout,
		// which opens the door to Slowloris and other slow-client attacks.
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,

		// Funnel http.Server's internal error log through slog so all logs
		// are in the same structured format.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// ── Start ────────────────────────────────────────────────────────────────
	// Run the server in a goroutine so it does not block the signal handler below.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// ── Graceful shutdown ────────────────────────────────────────────────────
	// Block until we receive SIGINT (Ctrl-C) or SIGTERM (from a process manager
	// or container orchestrator like Docker / Kubernetes).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		// Server failed to start (e.g. port already in use).
		slog.Error("server error", "err", err)
		os.Exit(1)
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	// Give in-flight requests up to ShutdownTimeout to complete.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

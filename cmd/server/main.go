// Command server runs the Botanary API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devsecops-playground-org/botanary-backend/internal/api"
	"github.com/devsecops-playground-org/botanary-backend/internal/catalog"
	"github.com/devsecops-playground-org/botanary-backend/internal/config"
)

// revision is stamped at build time with the commit the image was built from.
var revision = "unknown"

func main() {
	cfg := config.Load()

	// The runtime image is distroless: no shell, no curl. The container
	// healthcheck therefore re-executes this binary with -healthcheck.
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(healthcheck(cfg.Port))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if missing := cfg.MissingProductionSecrets(); len(missing) > 0 {
		logger.Error("refusing to start", "missing", missing)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewServer(catalog.NewSeeded(), logger).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("listening", "addr", srv.Addr, "env", cfg.Env, "revision", revision)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	// Finish in-flight requests before exiting so a rolling deploy drops nothing.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("stopped")
}

func healthcheck(port string) int {
	client := &http.Client{Timeout: 4 * time.Second}

	resp, err := client.Get("http://127.0.0.1:" + port + "/health")
	if err != nil {
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
